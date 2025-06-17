# BibleBrain Services

A serverless application providing API services for BibleBrain, primarily focused on copyright management and PDF generation.

## Overview

This project is built with Go and deployed on AWS Lambda using the Serverless Framework. It provides API endpoints for copyright information retrieval and PDF document generation based on product codes.

## Prerequisites

- Go 1.24+
- Node.js and npm
- AWS CLI configured with appropriate permissions
- Docker (for local development)
- MySQL database access

## Development Setup

This project uses a development container with all necessary tools pre-installed:

1. Clone the repository:
   ```sh
   git clone <repository-url>
   cd biblebrain-services
   ```

2. Configure environment variables:
   ```sh
   cp .devcontainer/.env.template .devcontainer/.env
   # Edit .env file with your configuration
   ```

3. Start the development container:
   ```sh
   # Using VS Code
   # Open the project in Visual Studio Code and click "Reopen in Container"
   
   # Or use Docker Compose directly
   docker-compose -f .devcontainer/docker-compose.yml up -d
   ```

4. Install dependencies:
   ```sh
   go mod tidy
   npm install
   ```

## Available Commands

The project uses a Makefile for common operations:

```sh
# Build the application
make build

# Clean build artifacts
make clean

# Run serverless offline for local testing
make offline

# Deploy to AWS
make deploy

# Format and standardize code
make standardize

# Run pre-commit checks (formatting and linting)
make precommit
```

## API Endpoints

### Status Endpoint

- **Path**: `/api/status`
- **Method**: GET
- **Parameters**: `name` (query string)
- **Description**: Returns a simple status message to verify the service is running

### Copyright Creation Endpoint

- **Path**: `/api/copyright/create`
- **Method**: POST
- **Request Body**: JSON with array of products
  ```json
  {
    "productList": [
      {"ProductCode": "P1PUI/LAN"},
      {"ProductCode": "N2ENG/NIV"}
    ]
  }
  ```
- **Response**: PDF document containing copyright information
- **Content-Type**: application/pdf

## Environment Configuration

The application uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (Debug, Error, Info) | Debug |
| `BIBLEBRAIN_DSN` | Database connection string for local development | - |
| `BIBLEBRAIN_DSN_SSM_ID` | SSM parameter ID for database connection in AWS | /dev/biblebrain/sql/dsn-otc00l0j3b9ggbgc |
| `environment` | Deployment environment (local, dev, prod) | local |

## Deployment

The application is deployed using Serverless Framework:

```sh
# Deploy to development environment
sls deploy --stage dev

# Deploy to production environment
sls deploy --stage prod
```

## Testing

Run tests with the standard Go test command:

```sh
go test ./...
```

Integration tests for the copyright service are available in copyright_test.go.

## Project Structure

```
.
├── bin/                   # Compiled binaries
├── cmd/                   # Command entry points
│   └── httpserver/        # HTTP server implementation
│       └── api/           # API handlers
├── service/               # Business logic services
│   ├── connection/        # Database connection handling
│   ├── copyright/         # Copyright service implementation
│   ├── pdf/               # PDF generation utilities
│   └── sign/              # AWS signature utilities
├── util/                  # Utility functions
├── .devcontainer/         # Development container configuration
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── package.json           # Node.js dependencies
└── serverless.yml         # Serverless Framework configuration
```
