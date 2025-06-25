package service

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"web-crawler/store/sqlc"

	"github.com/chromedp/chromedp"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"
)

type CreateEmasParams struct {
	CreatedAt time.Time
}

type CreateEmasResult struct {
	ID int32
}

// crawlGoldPrices fetches gold prices from Pegadaian website using headless browser
// This method handles JavaScript-rendered content properly
func (service *Service) crawlGoldPrices(ctx context.Context) (float64, float64, error) {
	const op = "[service] - Service.crawlGoldPrices"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.WithFields(logrus.Fields{
		"message": "Starting gold price crawling using headless browser",
	}).Info()

	// Create a new browser context with timeout
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(logger.Printf))
	defer cancel()

	// Set a reasonable timeout for the entire operation
	ctx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()

	var pageContent string
	var priceElements []string

	err := chromedp.Run(ctx,
		// Navigate to the gold price page
		chromedp.Navigate("https://sahabat.pegadaian.co.id/harga-emas"),

		// Wait for the page to load
		chromedp.WaitVisible("body", chromedp.ByQuery),

		// Wait a bit more for JavaScript to load dynamic content
		chromedp.Sleep(5*time.Second),

		// Get the full page content for debugging
		chromedp.InnerHTML("html", &pageContent, chromedp.ByQuery),

		// Try to find elements containing gold price patterns
		chromedp.Evaluate(`
			(function() {
				// Look for any text containing price patterns
				const elements = Array.from(document.querySelectorAll('*'));
				const pricePatterns = ['0,01 gr', '0.01 gr', '/ 0,01 gr', '/ 0.01 gr'];
				const foundElements = [];
				
				elements.forEach(el => {
					const text = el.textContent || '';
					pricePatterns.forEach(pattern => {
						if (text.includes(pattern) && text.length < 200) {
							foundElements.push(text.trim());
						}
					});
				});
				
				// Remove duplicates and return
				return [...new Set(foundElements)];
			})()
		`, &priceElements),
	)

	if err != nil {
		err = fmt.Errorf("failed to scrape website with headless browser: %w", err)

		logger.WithError(err).Error()

		return 0, 0, err
	}

	logger.WithFields(logrus.Fields{
		"found":              len(priceElements),
		"potential_elements": fmt.Sprintf("%+v", priceElements),
	}).Info()

	// Extract prices from the found elements
	var prices []float64

	// Look for price patterns in the extracted elements
	// The format appears to be: "Rp 18.500 / 0,01 gr" with periods as thousands separators
	patterns := []string{
		"/ 0,01 gr",
		"/ 0.01 gr",
		"/0,01 gr",
		"/0.01 gr",
		"0,01 gr",
		"0.01 gr",
	}

	for _, element := range priceElements {
		// Clean up Unicode characters that might interfere with regex matching
		cleanElement := strings.ReplaceAll(element, "\u00a0", " ") // Replace non-breaking space with regular space
		cleanElement = strings.ReplaceAll(cleanElement, "  ", " ") // Replace multiple spaces with single space
		cleanElement = strings.TrimSpace(cleanElement)

		for _, pattern := range patterns {
			if strings.Contains(cleanElement, pattern) {
				// Extract price using regex - handle both comma and period as decimal/thousand separators
				// Look for patterns like "Rp 18.500 / 0,01 gr" or "18.500 / 0,01 gr"
				re := regexp.MustCompile(`(?i)(?:rp\s*)?([0-9]+(?:\.[0-9]{3})*(?:,[0-9]+)?)\s*` + regexp.QuoteMeta(pattern))
				matches := re.FindStringSubmatch(cleanElement)

				if len(matches) > 1 {
					priceStr := matches[1]
					// Handle Indonesian number format: periods for thousands, comma for decimal
					// Convert "18.500" to "18500" and "18.500,50" to "18500.50"
					parts := strings.Split(priceStr, ",")
					integerPart := strings.ReplaceAll(parts[0], ".", "")

					var finalPriceStr string
					if len(parts) > 1 {
						// Has decimal part
						finalPriceStr = integerPart + "." + parts[1]
					} else {
						// No decimal part
						finalPriceStr = integerPart
					}

					if price, err := strconv.ParseFloat(finalPriceStr, 64); err == nil {
						// Multiply by 100 since prices are per 0.01 gr to get price per gram
						finalPrice := price * 100

						prices = append(prices, finalPrice)
					} else {
						logger.WithFields(logrus.Fields{
							"raw_text":        element,
							"clean_text":      cleanElement,
							"price_str":       priceStr,
							"final_price_str": finalPriceStr,
							"error":           err,
						}).Warn("Failed to parse extracted price")
					}
				}
			}
		}
	}

	// If we couldn't find prices in structured elements, try a broader search
	if len(prices) < 2 {
		logger.Warn("Could not find prices in structured elements, trying broader search")

		// Search the entire page content for price patterns
		// Clean up Unicode characters in page content
		cleanPageContent := strings.ReplaceAll(pageContent, "\u00a0", " ")

		for _, pattern := range patterns {
			re := regexp.MustCompile(`(?i)(?:rp\s*)?([0-9]+(?:\.[0-9]{3})*(?:,[0-9]+)?)\s*` + regexp.QuoteMeta(pattern))
			matches := re.FindAllStringSubmatch(cleanPageContent, -1)

			for _, match := range matches {
				if len(match) > 1 {
					priceStr := match[1]
					// Handle Indonesian number format: periods for thousands, comma for decimal
					parts := strings.Split(priceStr, ",")
					integerPart := strings.ReplaceAll(parts[0], ".", "")

					var finalPriceStr string
					if len(parts) > 1 {
						// Has decimal part
						finalPriceStr = integerPart + "." + parts[1]
					} else {
						// No decimal part
						finalPriceStr = integerPart
					}

					if price, err := strconv.ParseFloat(finalPriceStr, 64); err == nil {
						// Multiply by 100 since prices are per 0.01 gr to get price per gram
						finalPrice := price * 100

						prices = append(prices, finalPrice)
					} else {
						logger.WithFields(logrus.Fields{
							"pattern":           pattern,
							"matched_price_str": priceStr,
							"parsed_price":      price,
							"final_price_str":   finalPriceStr,
							"error":             err,
						}).Warn("Failed to parse extracted price")
					}
				}
			}
		}
	}

	if len(prices) < 2 {
		err = fmt.Errorf("could not find both gold prices on the website, found %d prices", len(prices))

		logger.WithError(err).Error()

		// Log a sample of the page content for debugging
		contentSample := pageContent
		if len(contentSample) > 1000 {
			contentSample = contentSample[:1000] + "..."
		}

		logger.WithField("page_sample", contentSample).Debug("Page content sample")

		return 0, 0, err
	}

	// Remove duplicate prices
	uniquePrices := make(map[float64]bool)
	var distinctPrices []float64
	for _, price := range prices {
		if !uniquePrices[price] {
			uniquePrices[price] = true
			distinctPrices = append(distinctPrices, price)
		}
	}

	if len(distinctPrices) < 2 {
		err = fmt.Errorf("could not find two distinct gold prices on the website, found %d distinct prices", len(distinctPrices))

		logger.WithError(err).Error()

		return 0, 0, err
	}

	// Sort distinct prices to determine which is higher (Jual) and which is lower (Beli)
	var jual, beli float64
	if distinctPrices[0] > distinctPrices[1] {
		jual = distinctPrices[0] // Higher price is selling price
		beli = distinctPrices[1] // Lower price is buying price
	} else {
		jual = distinctPrices[1] // Higher price is selling price
		beli = distinctPrices[0] // Lower price is buying price
	}

	logger.WithFields(logrus.Fields{
		"jual_price":       jual,
		"beli_price":       beli,
		"price_difference": jual - beli,
	}).Info()

	return jual, beli, nil
}

func (service *Service) CreateEmas(ctx context.Context, params *CreateEmasParams) (*CreateEmasResult, error) {
	const op = "[service] - Service.CreateEmas"

	// Initialize logger
	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize result
	result := &CreateEmasResult{}

	// Crawl gold prices from website
	jual, beli, err := service.crawlGoldPrices(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to crawl gold prices")
		return nil, fmt.Errorf("failed to crawl gold prices: %w", err)
	}

	// Create emas
	emas, err := service.store.CreateEmas(ctx, sqlc.CreateEmasParams{
		Jual: pgtype.Numeric{
			Int:   big.NewInt(int64(jual)),
			Valid: true,
		},
		Beli: pgtype.Numeric{
			Int:   big.NewInt(int64(beli)),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamp{
			Time:  params.CreatedAt,
			Valid: true,
		},
	})
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	// Set result
	result.ID = emas.EmasID.Int32

	return result, nil
}
