# Core Backend Service

A standard Go/Gin backend service following best practices and clean architecture principles.

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go        # Application entry point
├── internal/              # Private application code
│   ├── handler/           # HTTP handlers (Controller layer)
│   ├── service/           # Business logic layer
│   ├── repository/        # Data access layer (DAO/DB)
│   ├── model/             # Database models and domain entities
│   └── config/            # Configuration structures
├── pkg/                   # Public library code
│   ├── logger/            # Custom logger
│   └── utils/             # Utility functions
├── api/                   # API protocol definitions
│   ├── openapi/           # Swagger/OpenAPI specifications
│   └── proto/             # gRPC .proto files
├── configs/               # Configuration files (yaml, json, toml)
├── scripts/               # Build, install, and analysis scripts
├── build/                 # Packaging and CI
│   ├── package/           # Dockerfiles
│   └── ci/                # CI configuration files
├── deployments/           # Deployment configurations (K8s, Helm)
├── test/                  # External test data and integration tests
├── go.mod                 # Dependency management
├── go.sum
├── Makefile               # Common commands
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker (optional, for containerization)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   make deps
   ```

### Running the Application

```bash
# Run directly
make run

# Or build and run
make build
./build/bin/core
```

### Development

```bash
# Format code
make fmt

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

### Database Migrations

`AutoMigrate` is controlled by `database.auto_migrate` in `configs/config.yaml` and should stay `false` in production.

Use SQL migrations managed by Goose:

```bash
# install goose (one-time)
go install github.com/pressly/goose/v3/cmd/goose@latest

# create a new migration file template (no DB change yet)
make migrate-create name=add_coupon_scope_fields
# then edit apps/core/migrations/<new>.sql (-- +goose Up / -- +goose Down)

# apply all migrations
make migrate-up DB_DSN='postgres://postgres:postgres@localhost:5432/revieu?sslmode=disable'

# check status
make migrate-status DB_DSN='postgres://postgres:postgres@localhost:5432/revieu?sslmode=disable'
```

`make migrate-create` only generates the SQL migration file. The database schema changes when `make migrate-up` runs.

### Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## Configuration

Configuration files should be placed in the `configs/` directory. The application supports multiple configuration formats (YAML, JSON, TOML).

## API Documentation

API documentation is available in the `api/openapi/` directory.

## License

[Your License Here]
