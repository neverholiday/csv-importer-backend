# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CSV importer backend service that provides REST APIs for:
- Creating events with CSV file uploads
- Listing events
- Processing todo items from CSV files

The application uses Echo web framework, GORM for database operations, and PostgreSQL as the database.

## Development Setup

1. Create `.env` file:
```sh
CSV_IMPORTER_DB_HOST=localhost
CSV_IMPORTER_DB_PORT=5432
CSV_IMPORTER_DB_USER=postgres
CSV_IMPORTER_DB_PASSWORD=mypassword
CSV_IMPORTER_DB_NAME=postgres
```

2. Start PostgreSQL database:
```sh
docker compose up -d
```

3. Source environment variables:
```sh
export $(grep -v '^#' .env | xargs)
```

4. Run the application:
```sh
go run ./cmd/csv-importer
```

## Common Commands

- **Run application**: `go run ./cmd/csv-importer`
- **Build application**: `go build ./cmd/csv-importer`
- **Start database**: `docker compose up -d`
- **Stop database**: `docker compose down`
- **View database logs**: `docker logs postgres`

## Architecture

### Project Structure
- `cmd/csv-importer/` - Main application entry point and business logic
  - `main.go` - Application bootstrap, database connection, and Echo server setup
  - `apis/` - HTTP handlers and routing
  - `model/` - Data models and request/response structures
  - `repository/` - Database operations using GORM
- `sql/` - Database initialization scripts
- `docker-compose.yml` - PostgreSQL container setup

### Key Components

1. **Database Models** (`model/event.go`):
   - `Event` - Main events table with status tracking (draft/start/end)
   - `TodoEvent` - Related todo items for events
   - `TodoCSV` - CSV parsing structure for todo items

2. **API Layer** (`apis/event.go`):
   - `GET /api/v1/events` - List all events
   - `POST /api/v1/event` - Create event with CSV file upload
   - Uses multipart form data for file uploads

3. **Repository Layer** (`repository/event.go`):
   - GORM-based database operations
   - Context-aware database queries
   - Debug mode enabled for SQL logging

4. **Configuration**:
   - Environment-based configuration using `envconfig`
   - Database connection parameters prefixed with `CSV_IMPORTER_`
   - UTC timezone enforcement

### Technology Stack
- **Framework**: Echo v4 for HTTP server
- **ORM**: GORM with PostgreSQL driver
- **Database**: PostgreSQL 15
- **CSV Processing**: gocarina/gocsv
- **UUID Generation**: google/uuid (V7)
- **Configuration**: kelseyhightower/envconfig

The application runs on port 8080 and expects PostgreSQL to be available on the configured database parameters.