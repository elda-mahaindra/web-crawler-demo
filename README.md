# Web Crawler Demo

A demonstration project for web crawling and scraping that extracts gold prices from Pegadaian's website and stores them in a PostgreSQL database.

## Author

Elda Mahaindra ([faith030@gmail.com](mailto:faith030@gmail.com))

## Overview

This project demonstrates website crawling/scraping techniques to extract specific values and perform actions as needed. Specifically, it:

- Scrapes gold buying and selling prices from `https://sahabat.pegadaian.co.id/harga-emas`
- Uses a headless Chrome browser via ChromeDP for JavaScript-rendered content
- Stores extracted data in PostgreSQL database
- Runs on a configurable schedule using a ticker-based scheduler
- Provides a complete containerized environment with Docker Compose

## Architecture

The project follows a clean architecture pattern with the following layers:

- **CMD Layer**: Application entry point and command handling
- **Service Layer**: Business logic and web scraping functionality  
- **Store Layer**: Database operations using SQLC for type-safe queries
- **Scheduler Layer**: Automated task scheduling and execution
- **Utilities**: Configuration management

### Key Technologies

- **Go 1.23** - Main programming language
- **ChromeDP** - Headless browser automation for scraping JavaScript-rendered content
- **PostgreSQL** - Database for storing gold price data
- **Docker & Docker Compose** - Containerization and orchestration
- **SQLC** - Type-safe SQL query generation
- **Logrus** - Structured logging
- **Viper** - Configuration management

## Prerequisites

- Docker and Docker Compose
- Make (for convenience commands)

## Setup and Installation

### 1. Clone the Repository

```bash
git clone https://github.com/elda-mahaindra/web-crawler-demo.git
cd web-crawler-demo
```

### 2. Configure the Application

**Important:** Copy the sample configuration file and rename it before running:

```bash
cp web-crawler/config.json.sample web-crawler/config.json
```

### 3. Run the Application

Start the development environment:

```bash
make dev-up
```

Stop the application:

```bash
make dev-down
```

Stop and remove volumes (fresh start):

```bash
make dev-down-v
```

## Configuration

The application uses a JSON configuration file (`web-crawler/config.json`). Here's what each section configures:

### Configuration Structure

```json
{
  "app": {
    "name": "web-crawler",
    "host": "0.0.0.0",
    "port": 4000
  },
  "db": {
    "postgres": {
      "connection_string": "postgresql://postgres:changeme@postgres:5432/web_crawler_demo_db",
      "pool": {
        "max_conns": 25,
        "min_conns": 5
      }
    }
  },
  "scheduler": {
    "setups": [
      {
        "id": "gold_price",
        "url": "https://sahabat.pegadaian.co.id/harga-emas",
        "ticker_duration": "24h",
        "retry": {
          "max_attempts": 3,
          "initial_delay": "2s",
          "max_delay": "30s",
          "backoff_factor": 2.0,
          "enable_jitter": true
        }
      }
    ]
  }
}
```

### Configuration Sections Explained

#### App Section
- **name**: Application service name for identification
- **host**: Server host address (default: "0.0.0.0" to bind to all interfaces)
- **port**: Server port number (default: 4000) 

#### Database Section  
- **connection_string**: PostgreSQL connection string with credentials and database name
- **pool.max_conns**: Maximum number of database connections (default: 25)
- **pool.min_conns**: Minimum number of database connections (default: 5)

#### Scheduler Section
- **setups**: Array of scheduled tasks
  - **id**: Unique identifier for the scheduled task
  - **url**: Target website URL to scrape (e.g., "https://sahabat.pegadaian.co.id/harga-emas")
  - **ticker_duration**: How often to run the task (e.g., "24h" = every 24 hours)
  - **retry**: Retry configuration for handling scraping failures
    - **max_attempts**: Maximum number of retry attempts (default: 3)
    - **initial_delay**: Initial delay before first retry (e.g., "2s")
    - **max_delay**: Maximum delay between retries (e.g., "30s")
    - **backoff_factor**: Multiplier for exponential backoff (e.g., 2.0 means 2s, 4s, 8s...)
    - **enable_jitter**: Add randomization to delays to prevent thundering herd (default: true)

**Note for Development/Debugging:** While the sample configuration sets the scheduler to run every 24 hours, for debugging and testing purposes, it's recommended to use a shorter interval like `"10s"` (10 seconds) to see results quickly.

## How It Works

### 1. Web Scraping Process

The application uses ChromeDP (headless Chrome) to:
- Navigate to the Pegadaian gold price page
- Wait for JavaScript content to load
- Extract gold price elements containing patterns like "0,01 gr" or "0.01 gr"
- Parse Indonesian number format (periods as thousands separators, commas as decimal)
- Convert prices from per-0.01-gram to per-gram by multiplying by 100
- Identify buying (beli) and selling (jual) prices

### 2. Data Storage

Extracted prices are stored in PostgreSQL with the following schema:

```sql
CREATE TABLE ibdwh.emas (
    emas_id VARCHAR(10) PRIMARY KEY,  -- Date format: YYYY-MM-DD
    jual numeric NULL,                -- Selling price
    beli numeric NULL,                -- Buying price  
    created_at timestamp NULL,
    avg_bpkh numeric NULL
);
```

**Key Features:**
- **Date-based Primary Key**: `emas_id` uses YYYY-MM-DD format ensuring one record per day
- **UPSERT Logic**: Uses `ON CONFLICT` to update existing records if prices change during the day
- **Data Integrity**: Guarantees exactly one price record per date
- **Idempotent Operations**: Safe to run multiple times without creating duplicates

### 3. Scheduling

The scheduler runs automatically based on configuration:
- Configurable interval between executions (default: 24 hours)
- Continuous monitoring with proper error handling

### 4. REST API

The application provides a REST API for accessing scraped data:
- Built with Fiber framework for high performance
- Supports pagination for large datasets
- Returns JSON responses with structured data

**Primary Purpose**: The REST API serves as the main way to verify that the gold price scraping is working correctly without requiring direct database access. Instead of connecting to PostgreSQL manually, users can simply call the API endpoints to see the collected data and confirm the demo is functioning properly.

### 5. Retry Mechanism

To ensure data collection reliability, the application implements a robust retry mechanism:

- **Exponential Backoff**: Delays between retries increase exponentially (e.g., 2s, 4s, 8s, 16s)
- **Configurable Attempts**: Set maximum number of retry attempts per scraping operation
- **Maximum Delay Cap**: Prevents delays from becoming too long
- **Jitter**: Adds randomization to delays to prevent multiple instances from overwhelming the server
- **Context Cancellation**: Respects context cancellation during retry waits
- **Detailed Logging**: Logs each attempt with timing and error details for debugging

This mechanism helps handle temporary issues like:
- Network connectivity problems
- Server overload at the target website
- Rate limiting responses
- Temporary DNS resolution failures

**Example retry sequence with default config:**
1. First attempt fails → Wait 2s
2. Second attempt fails → Wait 4s  
3. Third attempt fails → Operation fails with detailed error

## Development

### Project Structure

```
web-crawler-demo/
├── _init/postgres/          # Database initialization scripts
├── docs/
│   └── postman/             # Postman collection for API testing
├── web-crawler/
│   ├── cmd/                 # Application commands
│   ├── api/                 # REST API endpoints
│   ├── middleware/          # HTTP middleware
│   ├── scheduler/           # Task scheduling logic
│   ├── service/             # Business logic and scraping
│   ├── store/               # Database layer
│   │   ├── queries/         # SQL queries
│   │   ├── schemas/         # Database schemas  
│   │   └── sqlc/            # Generated type-safe queries
│   └── util/                # Utilities (config, etc.)
├── docker-compose.dev.yml   # Development environment
└── Makefile                 # Convenience commands
```

### Available Make Commands

- `make dev-up` - Start development environment in detached mode
- `make dev-down` - Stop development environment  
- `make dev-down-v` - Stop development environment and remove volumes
- `make help` - Show available commands

### Application Commands

Inside the container, the application supports:

```bash
# Show help
./web-crawler help

# Start the service
./web-crawler start
```

## Verifying the Demo Works

After starting the application, you can verify that the gold price scraping is working through the REST API:

### Available API Endpoints

- **GET /emas** - List all gold price records with pagination
  - Query parameters:
    - `page` (optional): Page number (default: 1)
    - `size` (optional): Records per page (default: 10)

### Example API Usage

```bash
# Get all gold price records (first page, 10 records)
curl "http://localhost:4000/emas?page=1&size=10"

# Get specific page
curl "http://localhost:4000/emas?page=2&size=5"
```

### Expected Response Format

```json
{
  "emas": [
    {
      "emas_id": "2025-06-25",
      "jual": 1850000,
      "beli": 1785000,
      "created_at": "2025-06-25T02:13:53.98117",
      "avg_bpkh": null
    }
  ],
  "page": 1,
  "size": 10,
  "pages": 1,
  "total": 4
}
```

### Using Postman Collection

For easier testing, we've provided a Postman collection:

1. Import the collection: `docs/postman/web-crawler-demo.postman_collection.json`
2. Set environment variables:
   - `host`: `localhost`
   - `port`: `4000`
3. Run the "list" request to see scraped gold prices

The collection includes sample requests and responses for reference.

**Note**: This is the easiest way to verify the demo is working without needing to connect directly to the PostgreSQL database. Simply run the API calls to see if gold prices are being scraped and stored successfully.

## Logging

The application provides structured logging with:
- JSON and text format options
- Different log levels (Debug, Info, Warn, Error)
- Contextual fields for debugging
- Operation tracking for traceability

## Database

The PostgreSQL database is automatically initialized with:
- Database: `web_crawler_demo_db`
- Schema: `ibdwh` 
- Table: `emas` for storing gold price data
- Credentials: `postgres/changeme` (configurable)

## Troubleshooting

### Common Issues

1. **Config file not found**: Ensure you've copied `config.json.sample` to `config.json`
2. **Database connection failed**: Check if PostgreSQL container is running and credentials are correct
3. **Scraping failed**: The target website may have changed structure or blocked requests
4. **Port conflicts**: Ensure ports 4000 (API) and 5432 (PostgreSQL) are available

### Logs

View application logs:
```bash
docker logs web-crawler
```

View database logs:
```bash
docker logs postgres
```

## License

This project is for demonstration and educational purposes. Feel free to use it as a reference for implementing web crawling and scraping techniques in your own applications.

---

_Built with ❤️ to help developers understand web scraping, containerization, and clean architecture patterns_

