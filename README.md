# App Template

A flexible Go application template framework that supports multiple service types and optional component integration, suitable for quickly building various backend services.

## Features

- **Modular design**: Enable only what you need
- **Dependency injection**: Using Uber Fx
- **Configuration management**: YAML/JSON with environment variable override for sensitive data only
- **Structured logging**: Zap with log rotation
- **Graceful shutdown**: Signal handling and resource cleanup
- **Health checks**: /health, /readiness, /liveness endpoints
- **CLI tools**: Built with Cobra
- **Database integration**: Bun ORM with support for MySQL, PostgreSQL, SQLite
- **Database migrations**: Built-in migration tool for schema management
- **Code generation**: Model generation from database tables
- **Observability**: Configurable distributed tracing (OpenTelemetry) and metrics (Prometheus)

## Quick Start

### Prerequisites

- Go 1.26.4 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/ilaziness/orange-tv.git
cd orange-tv

# Download dependencies
make deps
```

### Running the Application

```bash
# Run with default configuration
make run

# Run with development configuration
make run-dev

# Or build and run
make build
./build/orange-tv serve
```

### Available Commands

```bash
# Start the server with default config
./build/orange-tv serve

# Start with specific environment (dev/prod/test)
./build/orange-tv serve -e dev
./build/orange-tv serve -e prod
./build/orange-tv serve -e test

# Start with specific config file
./build/orange-tv serve -c configs/config.prod.yaml

# Show version
./build/orange-tv version

# Validate configuration
./build/orange-tv config validate -e dev
./build/orange-tv config validate -c configs/config.prod.yaml

# Show current configuration
./build/orange-tv config show -e dev
./build/orange-tv config show -c configs/config.prod.yaml

# Database migrations
./build/orange-tv migrate up              # Run all pending migrations
./build/orange-tv migrate down            # Rollback last migration
./build/orange-tv migrate status          # Show migration status
./build/orange-tv migrate create <name>   # Create new migration files
./build/orange-tv migrate up --dry-run    # Preview migrations without executing

# Code generation
./build/orange-tv gen model               # Generate models from database
./build/orange-tv gen model --table users --output ./internal/model

# Show help
./build/orange-tv --help
```

## Configuration

Configuration can be provided via:

1. **YAML files**: `configs/config.yaml` (base), `config.dev.yaml`, `config.prod.yaml`, `config.test.yaml`
2. **Command-line**: `--env dev|prod|test` or `--config <path>`
3. **Environment variables**: Only for sensitive data (passwords, API keys)

### Configuration Priority

`--config` > `--env` > `config.yaml` defaults

### Configuration Files

- `config.yaml` - Base configuration
- `config.dev.yaml` - Development environment (use `--env dev`)
- `config.prod.yaml` - Production environment (use `--env prod`)
- `config.test.yaml` - Test environment (use `--env test`)

### Environment Variables

Environment variables are only used for sensitive data:

```bash
# Database configuration (MySQL default)
export DATABASE_DRIVER=mysql
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=3306
export DATABASE_USER=orange
export DATABASE_PASSWORD=orange_password
export DATABASE_DATABASE=orange_tv

# Redis configuration
export REDIS_ENABLED=false
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
```

See `.env.example` for all available environment variables.

## Database Integration

This project uses Bun ORM with **MySQL as the default** for development and production. PostgreSQL and SQLite remain supported by the driver layer (SQLite is still used in unit tests).

### Supported Databases

- **MySQL**: Default for development and production
- **PostgreSQL**: Optional
- **SQLite**: Optional / unit tests

### Database Configuration

Configure the database in your config files:

```yaml
database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: orange_tv
  user: orange
  password: orange_password
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 300
```

### Running Migrations

```bash
# Create a new migration
./build/orange-tv migrate create add_users_table

# Run migrations
./build/orange-tv migrate up

# Check migration status
./build/orange-tv migrate status

# Rollback last migration
./build/orange-tv migrate down
```

### Generating Models

```bash
# Generate models from all tables
./build/orange-tv gen model

# Generate model for specific table
./build/orange-tv gen model --table users

# Customize output
./build/orange-tv gen model --table users --output ./internal/model --package model
```

## Project Structure

```
.
├── cmd/                    # Command line interface
│   ├── root.go            # Root command
│   ├── serve.go           # Serve command
│   ├── config.go          # Config management commands
│   ├── version.go         # Version command
│   ├── migrate.go         # Database migration commands
│   └── gen.go             # Code generation commands
├── configs/               # Configuration files
├── migrations/            # Database migration files
│   ├── *.up.sql          # Up migration files
│   └── *.down.sql        # Down migration files
├── internal/              # Private application code
│   ├── app/              # Application wiring and lifecycle
│   ├── config/           # Configuration structures
│   ├── constant/         # Application constants
│   ├── database/         # Database initialization
│   ├── errcode/          # Error codes
│   ├── handler/          # Request handlers
│   │   ├── http/         # HTTP handlers
│   │   │   ├── user.go   # User handler
│   │   │   └── health.go # Health check handler
│   ├── logger/           # Logging wrapper
│   ├── middleware/       # Middlewares
│   │   └── http/         # HTTP-specific middlewares
│   ├── response/         # API response structures
│   ├── router/           # Route registration
│   ├── server/           # Server implementations
│   │   └── http.go       # HTTP server
│   ├── service/          # Business logic layer
│   │   └── user.go       # User service
│   ├── repository/       # Data access layer
│   │   └── user.go       # User repository
│   ├── model/            # Data models
│   │   └── user.go       # User model
│   └── dto/              # Data transfer objects
│       └── user.go       # User DTOs
├── main.go               # Application entry point
├── Makefile              # Build commands
├── .env.example          # Environment variables example
└── README.md             # This file
```

## Error Codes

Error codes follow the format `{3-digit module code}{4-digit business code}`:

- `100xxxx` - General module (parameter errors, data not found, etc.)
- `200xxxx` - User module (user not found, user exists, etc.)
- `300xxxx` - Auth module (auth failed, token expired, permissions, etc.)
- `900xxxx` - System module (internal error, database error, cache error, etc.)

Examples:

- `1000001` - Parameter error
- `1000002` - Data not found
- `2000001` - User not found
- `3000001` - Authentication failed
- `3000003` - Insufficient permission
- `9000001` - Internal server error

Each error code includes an associated HTTP status code.

## Development

### Make Commands

```bash
make build          # Build the application
make run            # Run the application
make run-dev        # Run with development config
make test           # Run tests
make test-coverage  # Run tests with coverage
make clean          # Clean build artifacts
make deps           # Download dependencies
make lint           # Run linter
make fmt            # Format code
make vet            # Run go vet
```

### Health Check Endpoints

Once the server is running, you can access:

- `GET /health` - Basic health check
- `GET /readiness` - Readiness check (includes dependency checks)
- `GET /liveness` - Liveness check
- `GET /version` - Application version
- `GET /version` - Application version
- `GET /metrics` - Prometheus metrics (requires `metrics.enabled: true`)

Example:

```bash
curl http://localhost:8080/health
```

## Observability

The application supports configurable distributed tracing and metrics monitoring. See [Observability Guide](docs/observability.md) for detailed documentation.

### Features

- **Distributed Tracing**: OpenTelemetry with OTLP protocol (supports Jaeger, Tempo, etc.)
- **Metrics Monitoring**: Prometheus native client with HTTP, database, Redis metrics
- **Data Correlation**: trace_id automatically injected into logs and HTTP response headers

### Quick Start

Enable observability in `configs/config.yaml`:

```yaml
# Enable distributed tracing
tracing:
  enabled: true
  endpoint: localhost:4317  # OTLP gRPC endpoint
  sample_rate: 1.0

# Enable metrics monitoring
metrics:
  enabled: true
  path: /metrics
  labels:
    env: dev
    version: "1.0.0"
```

Access metrics endpoint: http://localhost:8080/metrics

For detailed configuration and usage, see [Observability Guide](docs/observability.md).

## API Documentation

The application integrates Swagger API documentation. Access the following URLs to view:

- Swagger UI: http://localhost:8080/swagger/index.html

### Generate API Documentation

```bash
# Generate Swagger documentation
make swagger

# Clean generated documentation
make swagger-clean
```

## Example Code

The project includes multiple example code snippets demonstrating how to use various components:

- **HTTP Service**: Endpoints like `/api/v1/users/{id}` are complete HTTP service examples
- `internal/event/example_test.go` - Event system usage examples
- `internal/cache/example_test.go` - Cache usage examples

## Documentation

- [Module Usage Guide](docs/module-usage.md) - How to select and remove unnecessary service modules
- [Deployment Documentation](docs/deployment.md) - Single instance, multi-instance, and Docker deployment guides
- [Observability Guide](docs/observability.md) - Distributed tracing and metrics monitoring configuration

## License

MIT License
