# CSV Importer Backend

A Go-based REST API service for importing and managing CSV todo data with PostgreSQL database storage.

## 🚀 Features

- **REST API**: Create events and upload CSV files
- **CSV Processing**: Parse todo items from uploaded CSV files
- **Database Storage**: PostgreSQL with GORM ORM
- **Event Management**: Track events with status (draft/start/end)
- **Health Checks**: Built-in health monitoring endpoints
- **Comprehensive Testing**: Unit, integration, security, and performance tests

## 🏗️ Architecture

### Project Structure

```
cmd/csv-importer/
├── main.go                    # Application entry point
├── apis/                      # HTTP handlers and routing
│   ├── event.go              # Event API endpoints
│   ├── event_test.go         # API tests
│   └── healthcheck.go        # Health check endpoint
├── model/                     # Data models and structures
│   ├── event.go              # Event and TodoEvent models
│   ├── csv.go                # CSV parsing models
│   ├── rest.go               # REST API models
│   └── *_test.go             # Model tests
├── repository/                # Database operations
│   ├── event.go              # Event repository
│   └── event_test.go         # Repository tests
├── *_test.go                 # Integration, security, performance tests
sql/                          # Database scripts
testdata/                     # Test data files
docker-compose.yml            # PostgreSQL container setup
```

### Technology Stack

- **Framework**: Echo v4 for HTTP server
- **Database**: PostgreSQL 15 with GORM ORM
- **CSV Processing**: gocarina/gocsv library
- **Configuration**: Environment-based with envconfig
- **Testing**: testify framework with comprehensive test coverage
- **Containerization**: Docker Compose for local development

## 📋 Prerequisites

- Go 1.23.1 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (via Docker)

## 🛠️ Setup and Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd csv-importer-backend
```

### 2. Environment Configuration

Create a `.env` file in the root directory:

```bash
CSV_IMPORTER_DB_HOST=localhost
CSV_IMPORTER_DB_PORT=5432
CSV_IMPORTER_DB_USER=postgres
CSV_IMPORTER_DB_PASSWORD=mypassword
CSV_IMPORTER_DB_NAME=postgres
```

### 3. Start Database

```bash
# Start PostgreSQL container
docker compose up -d

# Verify database is running
docker logs postgres
```

### 4. Load Environment Variables

```bash
# Source environment variables
export $(grep -v '^#' .env | xargs)
```

### 5. Install Dependencies

```bash
# Download Go modules
go mod download
```

### 6. Run Application

```bash
# Start the server
go run ./cmd/csv-importer

# Server will start on port 8080
# Health check: http://localhost:8080/healthz
```

## 📡 API Endpoints

### Health Check

```bash
GET /healthz
```

**Response:**
```json
{
  "status": "ok",
  "database": "connected"
}
```

### List Events

```bash
GET /api/v1/events
```

**Response:**
```json
{
  "data": [
    {
      "id": "event-123",
      "name": "Sample Event",
      "status": "draft",
      "create_date": "2025-01-09T10:00:00Z",
      "update_date": "2025-01-09T10:00:00Z"
    }
  ],
  "message": "success"
}
```

### Create Event with CSV Upload

```bash
POST /api/v1/event
Content-Type: multipart/form-data
```

**Form Fields:**
- `name`: Event name (string)
- `csvfile`: CSV file with todo items

**CSV Format:**
```csv
todo_name,note
Buy groceries,Milk and bread
Call dentist,Schedule appointment
```

**Response:**
```json
{
  "data": {
    "event_id": "event-456",
    "todos_imported": 2
  },
  "message": "success"
}
```

## 🧪 Testing

The project includes a comprehensive test suite covering multiple aspects:

### Test Categories

- **Unit Tests**: Model validation, CSV parsing, configuration
- **Integration Tests**: Database operations with real PostgreSQL
- **API Tests**: HTTP endpoint testing with mocked dependencies
- **Security Tests**: Input validation, injection prevention, file upload safety
- **Performance Tests**: Concurrent operations, large data processing
- **Error Handling Tests**: Database failures, file system errors

### Running Tests

#### Run All Tests
```bash
# Execute complete test suite
go test -v ./cmd/csv-importer/

# Run tests with coverage
go test -cover ./cmd/csv-importer/

# Generate coverage report
go test -coverprofile=coverage.out ./cmd/csv-importer/
go tool cover -html=coverage.out
```

#### Run Specific Test Categories

```bash
# Model layer tests
go test -v ./cmd/csv-importer/model/

# Repository tests
go test -v ./cmd/csv-importer/repository/

# API tests  
go test -v ./cmd/csv-importer/apis/

# Main application tests
go test -v ./cmd/csv-importer/ -run TestEnvCfg

# CSV processing tests
go test -v ./cmd/csv-importer/ -run TestCSVProcessing

# Security tests
go test -v ./cmd/csv-importer/ -run TestSecurity

# Performance tests
go test -v ./cmd/csv-importer/ -run TestPerformance

# Error handling tests
go test -v ./cmd/csv-importer/ -run TestErrorHandling
```

#### Integration Tests

Integration tests require a real PostgreSQL database:

```bash
# Start test database
docker compose up -d

# Run integration tests
INTEGRATION_TEST=1 go test -v ./cmd/csv-importer/ -run TestIntegration

# Run with integration database setup
INTEGRATION_TEST=1 go test -v ./cmd/csv-importer/ -run TestIntegration_DatabaseOperations
```

#### Performance Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./cmd/csv-importer/

# Run specific benchmarks
go test -bench=BenchmarkRepository ./cmd/csv-importer/
go test -bench=BenchmarkCSV ./cmd/csv-importer/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./cmd/csv-importer/
```

#### Test with Race Detection

```bash
# Run tests with race condition detection
go test -race ./cmd/csv-importer/

# Run specific concurrent tests
go test -race -run TestPerformance_Concurrent ./cmd/csv-importer/
```

### Test Data

Test data files are located in `testdata/`:

```bash
testdata/
├── valid.csv           # Valid CSV with sample data
├── empty.csv          # CSV with headers only
└── malformed.csv      # Malformed CSV for error testing
```

## 🔧 Database Management

### Database Schema

The application uses GORM auto-migration. Tables are created automatically:

```sql
-- Events table
CREATE TABLE events (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    status VARCHAR NOT NULL,
    create_date TIMESTAMP NOT NULL,
    update_date TIMESTAMP NOT NULL,
    delete_date TIMESTAMP NULL
);

-- Todo events table  
CREATE TABLE todo_events (
    id VARCHAR PRIMARY KEY,
    event_id VARCHAR NOT NULL,
    create_date TIMESTAMP NOT NULL,
    update_date TIMESTAMP NOT NULL,
    delete_date TIMESTAMP NULL,
    FOREIGN KEY (event_id) REFERENCES events(id)
);
```

### Database Operations

```bash
# View database logs
docker logs postgres

# Connect to database
docker exec -it postgres psql -U postgres -d postgres

# Stop database
docker compose down

# Reset database (removes all data)
docker compose down -v
docker compose up -d
```

## 🔒 Security Features

The application includes several security measures:

- **Input Validation**: All user inputs are validated and sanitized
- **File Upload Security**: File type and size validation
- **SQL Injection Protection**: Parameterized queries via GORM
- **CSV Injection Prevention**: Dangerous formula detection
- **Path Traversal Protection**: Filename sanitization
- **Rate Limiting**: Configurable upload limits
- **Error Handling**: Secure error messages without sensitive data exposure

## 📈 Performance Features

- **Connection Pooling**: Optimized database connection management
- **Concurrent Processing**: Supports multiple simultaneous uploads
- **Large File Handling**: Efficient processing of large CSV files
- **Memory Management**: Controlled memory usage for large datasets
- **Response Caching**: Optimized response handling

## 🐛 Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check if database is running
docker ps

# Restart database
docker compose down && docker compose up -d

# Check environment variables
echo $CSV_IMPORTER_DB_HOST
```

#### Port Already in Use
```bash
# Check what's using port 8080
lsof -i :8080

# Kill process or change port in main.go
```

#### CSV Upload Errors
```bash
# Verify CSV format matches expected headers
cat testdata/valid.csv

# Check file permissions
ls -la testdata/
```

### Debugging

Enable debug logging:

```bash
# Run with verbose GORM logging
go run ./cmd/csv-importer 2>&1 | grep -E "(SQL|ERROR)"
```

## 🚀 Development

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (if available)
golangci-lint run

# Check for vulnerabilities
govulncheck ./...
```

### Building

```bash
# Build binary
go build -o csv-importer ./cmd/csv-importer

# Build with optimizations
go build -ldflags="-s -w" -o csv-importer ./cmd/csv-importer

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o csv-importer-linux ./cmd/csv-importer
```

## 📄 License

This project is licensed under the MIT License. See LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the complete test suite
6. Submit a pull request

### Development Guidelines

- Follow Go naming conventions
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting
- Use meaningful commit messages