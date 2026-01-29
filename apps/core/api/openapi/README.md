# API Documentation

## Test Endpoints

### GET /api/v1/test
Returns a test message to verify the API is working.

**Response:**
```json
{
  "message": "Hello, World!",
  "status": "success"
}
```

### POST /api/v1/test
Echoes back the received message.

**Request Body:**
```json
{
  "message": "Your message here"
}
```

**Response:**
```json
{
  "message": "Your message here",
  "status": "success"
}
```

## Swagger Documentation

After starting the server, you can access the interactive Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

## Generating Swagger Docs

To regenerate the Swagger documentation after making changes:
```bash
make swagger
```

Or manually:
```bash
swag init -g cmd/app/main.go -o docs
```
