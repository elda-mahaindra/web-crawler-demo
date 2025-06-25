# Stage 1: Build environment
# Using golang alpine image as the base for building the application
FROM golang:1.23-alpine AS build-env

# Set working directory for the build
WORKDIR /build

# Copy all files from the current directory to the working directory
COPY . .

# Download and verify dependencies
RUN go mod tidy
RUN go mod download

# Build the Go application
# CGO_ENABLED=0: Pure Go (no C dependencies)
# GOOS=linux: Target OS
# Compile all Go files in cmd directory into a single binary named 'main'
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/*.go

# Stage 2: Production environment
# Using minimal alpine image for the final container
FROM alpine:latest

# Install additional packages needed for the application including Chrome
RUN apk update && apk upgrade && \
     apk add --no-cache bash git openssh curl \
     # Install Chromium and dependencies for headless browser automation
     chromium \
     nss \
     freetype \
     freetype-dev \
     harfbuzz \
     ca-certificates \
     ttf-freefont \
     # Additional dependencies for Chrome to run properly
     && rm -rf /var/cache/apk/*

# Set Chrome/Chromium path for chromedp
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser

# Set the working directory for the application
WORKDIR /app

# Copy only necessary artifacts from the build stage
COPY --from=build-env /build/main main
COPY --from=build-env /build/config.json config.json

# Note: config.json can also be mounted via docker-compose volume
# This allows for different configs (local vs docker) without rebuilding

# Run the executable with 'start' argument
ENTRYPOINT [ "./main", "start" ]
