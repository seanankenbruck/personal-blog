# Personal Blog

A modern, test-driven personal blog built with Go, featuring OpenTelemetry instrumentation.

## Features

- RESTful API using Gin web framework
- PostgreSQL database for content storage
- OpenTelemetry instrumentation for metrics, logs, and traces
- Comprehensive test suite
- Clean architecture pattern

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Docker (optional, for development)

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   go mod tidy
   ```
3. Set up environment variables:
   ```bash
   cp .env.example .env
   ```
4. Update the `.env` file with your configuration
5. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Testing

Run the test suite:
```bash
go test ./...
```

## OpenTelemetry

The application is instrumented with OpenTelemetry. To collect and visualize telemetry data:

1. Set up an OpenTelemetry Collector
2. Configure the collector endpoint in your `.env` file
3. The application will automatically send metrics, logs, and traces to the collector

## Project Structure

```
.
├── cmd/                # Application entry points
├── internal/          # Private application code
│   ├── config/       # Configuration management
│   ├── domain/       # Domain models
│   ├── repository/   # Data access layer
│   ├── service/      # Business logic
│   └── transport/    # API handlers
├── pkg/              # Public packages
├── test/             # Test utilities and fixtures
└── docs/             # Documentation
```

## License

MIT