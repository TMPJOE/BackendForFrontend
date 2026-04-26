# BFF (Backend for Frontend) Service

A Go-based Backend for Frontend service that aggregates data from multiple downstream microservices (Hotel, Room, and Reservation services) to provide a unified API for frontend clients.

## Architecture

```
Client (Web/Mobile)
│
▼
API Gateway (auth + redirect, stays thin)
│
▼
┌─────────────────────────────────────────────────────────────┐
│ BFF Service (this)                                          │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │
│ │ Hotel       │ │ Room        │ │ Reservation         │    │
│ │ Client      │ │ Client      │ │ Client              │    │
│ └──────┬──────┘ └──────┬──────┘ └─────────┬─────────┘    │
│        │               │                  │              │
│ ┌──────▼───────────────▼──────────────────▼──────────┐    │
│ │ Service Layer (BFF)                                │    │
│ │ - Orchestrates calls to downstream services        │    │
│ │ - Aggregates responses                             │    │
│ │ - Handles business logic                           │    │
│ └──────┬─────────────────────────────────────────────┘    │
│        │                                                  │
│ ┌──────▼──────────────────────────────────────────────┐   │
│ │ Handler Layer (HTTP API)                              │   │
│ │ - Authentication (JWT-based)                          │   │
│ │ - Request validation                                  │   │
│ │ - Response formatting                                 │   │
│ └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
│
▼
Downstream Services
├── Hotel Service (port 8081)
├── Room Service (port 8082)
└── Reservation Service (port 8083)
```

## What is a BFF?

This service follows the **Backend for Frontend** pattern:

- **No direct database access** - it only communicates with downstream services via HTTP
- **Aggregates data** - combines data from multiple services into single responses
- **Validates relationships** - ensures hotels exist before creating rooms, rooms exist before creating reservations
- **Simplifies frontend** - frontend gets everything it needs in fewer API calls
- **Orchestrates workflows** - coordinates complex operations across multiple services

## Key Responsibilities

### 1. Data Aggregation
- `GET /hotels/{hotelId}/details` returns hotel + all rooms merged
- `GET /reservations/{reservationId}/details` returns reservation + hotel + room data merged

### 2. Cross-Service Validation
- `POST /hotels/{hotelId}/rooms` validates hotel exists before creating room
- `POST /reservations` validates hotel + room exist, calculates total amount, checks availability

### 3. Business Logic
- Calculates reservation total amounts based on room price × nights
- Validates check-in/check-out dates
- Enriches responses with hotel names and room numbers

## Tech Stack

- **Language**: Go 1.25
- **Router**: [go-chi/chi/v5](https://github.com/go-chi/chi)
- **Logging**: [go-chi/httplog/v3](https://github.com/go-chi/httplog) + `log/slog`
- **JWT Authentication**: [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)
- **Validation**: [go-playground/validator/v10](https://github.com/go-playground/validator)
- **Configuration**: YAML with environment variable expansion

## Features

### Security
- **JWT Authentication**: RSA-based token validation
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, HSTS
- **Request ID**: Unique request tracking for debugging and logging

### Resilience
- **Rate Limiting**: Token bucket algorithm with configurable limits
- **Request Timeouts**: Configurable per-service timeouts
- **Graceful Shutdown**: 30-second timeout for in-flight requests

## API Endpoints

### Public Routes (No Authentication)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Service health check |
| `GET` | `/ready` | Readiness check (validates downstream services) |

### Hotels (Authentication Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/hotels` | List all hotels (optional: `?city=&country=`) |
| `GET` | `/hotels/{hotelId}` | Get hotel by ID |
| `GET` | `/hotels/{hotelId}/details` | Get hotel with all its rooms |

### Rooms (Authentication Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/rooms/{roomId}` | Get room by ID |
| `POST` | `/hotels/{hotelId}/rooms` | Create a new room (validates hotel exists) |

### Reservations (Authentication Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/reservations` | List user's reservations |
| `GET` | `/reservations/{reservationId}` | Get reservation by ID |
| `GET` | `/reservations/{reservationId}/details` | Get full reservation details (hotel + room + reservation) |
| `POST` | `/reservations` | Create a new reservation (validates hotel + room, calculates total) |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HOTEL_SERVICE_URL` | Hotel Service URL | `http://localhost:8081` |
| `ROOM_SERVICE_URL` | Room Service URL | `http://localhost:8082` |
| `RESERVATION_SERVICE_URL` | Reservation Service URL | `http://localhost:8083` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |
| `LOG_FORMAT` | Logging format (json/text) | `json` |

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

jwt:
  issuer: "bff-service"
  expiration: "24h"

downstream_services:
  hotel_service_url: "${HOTEL_SERVICE_URL:http://localhost:8081}"
  room_service_url: "${ROOM_SERVICE_URL:http://localhost:8082}"
  reservation_service_url: "${RESERVATION_SERVICE_URL:http://localhost:8083}"
  timeout: 30s

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200

health:
  path: "/health"
  ready_path: "/ready"

logging:
  level: "info"
  format: "json"
```

## Getting Started

### Prerequisites

- Go 1.25+
- Downstream services running (Hotel, Room, Reservation)
- RSA key pair for JWT signing (`private.pem`, `public.pem`)

### Generate JWT Keys

```bash
# Generate private key
openssl genrsa -out private.pem 2048

# Generate public key
openssl rsa -in private.pem -pubout -out public.pem
```

### Run the Service

```bash
# Set downstream service URLs
export HOTEL_SERVICE_URL=http://localhost:8081
export ROOM_SERVICE_URL=http://localhost:8082
export RESERVATION_SERVICE_URL=http://localhost:8083

# Run the service
go run app/cmd/api/main.go
```

The server starts on `localhost:8080`.

### Test the Health Endpoint

```bash
# Health check
curl http://localhost:8080/health

# Readiness check (validates downstream services)
curl http://localhost:8080/ready
```

## Project Structure

```
app/
├── cmd/api/
│   └── main.go              # Application entry point
│
└── internal/
    ├── client/
    │   ├── client.go        # Base HTTP client
    │   ├── hotel_client.go  # Hotel Service client
    │   ├── room_client.go   # Room Service client
    │   ├── reservation_client.go  # Reservation Service client
    │   └── errors.go        # Client errors
    │
    ├── config/
    │   └── config.go        # Configuration loader
    │
    ├── handler/
    │   ├── handlers.go      # Base HTTP handlers
    │   ├── hotel_handlers.go  # Hotel route handlers
    │   ├── room_handlers.go   # Room route handlers
    │   ├── reservation_handlers.go  # Reservation route handlers
    │   ├── middleware.go    # Middleware (JWT, rate limiting, CORS)
    │   └── routing.go       # Route definitions
    │
    ├── helper/
    │   ├── util.go          # Utility functions
    │   └── validator.go     # Request validation
    │
    ├── logging/
    │   └── logger.go        # Structured logging
    │
    ├── models/
    │   └── models.go        # Domain models
    │
    └── service/
        ├── service.go       # Service interface
        ├── hotel_service.go # Hotel business logic
        ├── room_service.go  # Room business logic
        └── reservation_service.go  # Reservation business logic
```

## Downstream Service Contracts

This BFF expects downstream services to implement the following REST API:

### Hotel Service

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/hotels` | `[]Hotel` |
| `GET` | `/hotels/{id}` | `Hotel` |
| `POST` | `/hotels` | `Hotel` |
| `PUT` | `/hotels/{id}` | `Hotel` |
| `DELETE` | `/hotels/{id}` | `204` |
| `GET` | `/health` | `200` |

### Room Service

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/hotels/{hotelId}/rooms` | `[]Room` |
| `GET` | `/rooms/{id}` | `Room` |
| `POST` | `/rooms` | `Room` |
| `PUT` | `/rooms/{id}` | `Room` |
| `DELETE` | `/rooms/{id}` | `204` |
| `GET` | `/rooms/{id}/availability?check_in=&check_out=` | `bool` |
| `GET` | `/health` | `200` |

### Reservation Service

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/reservations` | `[]Reservation` |
| `GET` | `/reservations/{id}` | `Reservation` |
| `GET` | `/users/{userId}/reservations` | `[]Reservation` |
| `POST` | `/reservations` | `Reservation` |
| `PUT` | `/reservations/{id}` | `Reservation` |
| `PATCH` | `/reservations/{id}/status` | `Reservation` |
| `DELETE` | `/reservations/{id}` | `204` |
| `GET` | `/health` | `200` |

## Example Workflows

### Create a Room

```bash
curl -X POST http://localhost:8080/hotels/{hotelId}/rooms \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "room_number": "101",
    "type": "deluxe",
    "price": 200.00,
    "capacity": 2
  }'
```

### Create a Reservation

```bash
curl -X POST http://localhost:8080/reservations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "hotel_id": "{hotelId}",
    "room_id": "{roomId}",
    "guest_name": "John Doe",
    "guest_email": "john@example.com",
    "check_in": "2024-06-01T15:00:00Z",
    "check_out": "2024-06-05T11:00:00Z"
  }'
```

### Get Reservation Details

```bash
curl http://localhost:8080/reservations/{reservationId}/details \
  -H "Authorization: Bearer <token>"
```

Response:

```json
{
  "reservation": {
    "id": "res-123",
    "hotel_id": "hotel-456",
    "room_id": "room-789",
    "user_id": "user-abc",
    "guest_name": "John Doe",
    "guest_email": "john@example.com",
    "check_in": "2024-06-01T15:00:00Z",
    "check_out": "2024-06-05T11:00:00Z",
    "total_amount": 800.00,
    "status": "confirmed"
  },
  "hotel": {
    "id": "hotel-456",
    "name": "Grand Hotel",
    "city": "New York",
    "country": "USA"
  },
  "room": {
    "id": "room-789",
    "room_number": "101",
    "type": "deluxe",
    "price": 200.00
  }
}
```

## Error Handling

The BFF returns standard HTTP status codes:

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful, no content to return
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid JWT token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., room not available)
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Unexpected server error
- `503 Service Unavailable` - Downstream service unavailable

## Domain Models

### Hotel

```go
type Hotel struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Address     string    `json:"address"`
    City        string    `json:"city"`
    Country     string    `json:"country"`
    Rating      float64   `json:"rating"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Room

```go
type Room struct {
    ID          string    `json:"id"`
    HotelID     string    `json:"hotel_id"`
    HotelName   string    `json:"hotel_name,omitempty"`
    RoomNumber  string    `json:"room_number"`
    Type        string    `json:"type"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Capacity    int       `json:"capacity"`
    Amenities   []string  `json:"amenities"`
    IsAvailable bool      `json:"is_available"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Reservation

```go
type Reservation struct {
    ID          string    `json:"id"`
    HotelID     string    `json:"hotel_id"`
    HotelName   string    `json:"hotel_name,omitempty"`
    RoomID      string    `json:"room_id"`
    RoomNumber  string    `json:"room_number,omitempty"`
    UserID      string    `json:"user_id"`
    GuestName   string    `json:"guest_name"`
    GuestEmail  string    `json:"guest_email"`
    GuestPhone  string    `json:"guest_phone"`
    CheckIn     time.Time `json:"check_in"`
    CheckOut    time.Time `json:"check_out"`
    TotalAmount float64   `json:"total_amount"`
    Status      string    `json:"status"`
    Notes       string    `json:"notes"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o bff-service app/cmd/api/main.go
```

### Docker

```bash
# Build the image
docker build -t bff-service .

# Run with environment variables
docker run -p 8080:8080 \
  -e HOTEL_SERVICE_URL=http://hotel-service:8081 \
  -e ROOM_SERVICE_URL=http://room-service:8082 \
  -e RESERVATION_SERVICE_URL=http://reservation-service:8083 \
  -v /path/to/keys:/app/keys \
  bff-service
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f bff-service

# Stop services
docker-compose down
```
