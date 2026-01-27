# Quick Start Guide

## Installation

1. Install dependencies:
```bash
go mod tidy
```

2. Install Swagger CLI (if not already installed):
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

3. Generate Swagger documentation:
```bash
make swagger
```

## Running the Application

```bash
make run
```

The server will start on `http://localhost:8080` (or the address specified in your config).

## Testing the API

### Using curl

Test GET endpoint:
```bash
curl http://localhost:8080/api/v1/test
```

Test POST endpoint:
```bash
curl -X POST http://localhost:8080/api/v1/test \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from curl!"}'
```

### Using Swagger UI

Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

You can test all endpoints interactively through the Swagger UI.

## Available Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/test` - Test GET endpoint
- `POST /api/v1/test` - Test POST endpoint
- `GET /swagger/*any` - Swagger documentation UI

## Next Steps

1. Add your own handlers in `internal/handler/`
2. Add Swagger annotations to your handlers
3. Run `make swagger` to regenerate documentation
4. Test your endpoints via Swagger UI
