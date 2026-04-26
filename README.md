# 🔄 Backend for Frontend (BFF)

> Orchestration layer that aggregates and enriches data from multiple microservices for the frontend.

## Overview

The Backend-for-Frontend (BFF) is the **orchestration layer** between the frontend (Vue.js SPA) and the downstream microservices. It serves three key purposes:

1. **Data Aggregation** — Merges data from Hotel, Room, and Booking services into composite responses (e.g., hotel details + room list, reservation + hotel + room info)
2. **Service Bridging** — Validates cross-service constraints before forwarding (e.g., verifies hotel exists before creating a room in it)
3. **Business Enrichment** — Calculates derived fields like `total_price = room.price × nights` and injects `user_id` from JWT claims

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Router | [go-chi/chi](https://github.com/go-chi/chi) v5 |
| Auth | JWT verification (RSA-256 public key) |
| HTTP Clients | Custom typed clients for each downstream service |
| Container | Docker (multi-stage Alpine build) |

## Architecture

```
app/
├── cmd/api/          # Application entrypoint
│   └── main.go
├── internal/
│   ├── client/       # HTTP clients for downstream services
│   │   ├── client.go           # Base HTTP client
│   │   ├── errors.go           # Client error types
│   │   ├── hotel_client.go     # Hotel Service client
│   │   ├── room_client.go      # Room Service client
│   │   ├── reservation_client.go  # Legacy reservation client
│   │   ├── booking_client.go   # Booking Service client
│   │   └── payment_client.go   # Payment Service client
│   ├── config/       # YAML config loader with env var expansion
│   ├── handler/      # HTTP handlers, routing, JWT middleware
│   │   ├── handlers.go            # Base handler + health checks
│   │   ├── hotel_handlers.go      # Hotel aggregation endpoints
│   │   ├── room_handlers.go       # Room bridge endpoints
│   │   ├── reservation_handlers.go # Reservation orchestration
│   │   ├── middleware.go          # JWT, CORS, security, rate limit
│   │   └── routing.go            # Route definitions
│   ├── helper/       # Response helpers
│   ├── logging/      # Structured slog logger
│   ├── models/       # Aggregated domain models
│   │   └── models.go             # Hotel, Room, Booking, composite types
│   └── service/      # Business logic layer
│       ├── service.go            # Service interface + mappings
│       ├── hotel_service.go      # Hotel operations
│       ├── room_service.go       # Room operations
│       └── reservation_service.go # Reservation orchestration
├── config.yaml
├── Dockerfile
└── go.mod
```

## API Endpoints

All endpoints require JWT authentication (except health checks).

### Public Routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe (checks downstream services) |

### Hotel Aggregation Endpoints

| Method | Path | Type | Description |
|---|---|---|---|
| `GET` | `/hotels` | Passthrough | List hotels (forwarded to Hotel Service) |
| `GET` | `/hotels/{hotelId}` | Passthrough | Get hotel details |
| `GET` | `/hotels/{hotelId}/details` | **Aggregation** | Hotel + all its rooms (merged) |

### Room Bridge Endpoints

| Method | Path | Type | Description |
|---|---|---|---|
| `GET` | `/rooms/{roomId}` | Passthrough | Get room details |
| `POST` | `/hotels/{hotelId}/rooms` | **Bridge** | Verify hotel exists → create room |

### Reservation Orchestration Endpoints

| Method | Path | Type | Description |
|---|---|---|---|
| `GET` | `/reservations` | Passthrough | List user's reservations |
| `GET` | `/reservations/{id}` | Passthrough | Get reservation |
| `GET` | `/reservations/{id}/details` | **Aggregation** | Reservation + hotel + room (merged) |
| `POST` | `/reservations` | **Orchestration** | Full booking flow (validate → price → pay → book) |

## Flow Diagram

```mermaid
flowchart TD
    A["Frontend Request"] --> B["JWT Middleware"]
    B --> B1{"Token Valid?"}
    B1 -->|No| B2["401 Unauthorized"]
    B1 -->|Yes| C{"Endpoint Type?"}
    
    C -->|GET /hotels/id/details| D["AGGREGATION"]
    D --> D1["Fetch Hotel from Hotel Service"]
    D1 --> D2["Fetch Rooms from Room Service"]
    D2 --> D3["Map + Merge into HotelWithRooms"]
    D3 --> D4["Return composite JSON"]
    
    C -->|POST /hotels/id/rooms| E["BRIDGE"]
    E --> E1["Fetch Hotel from Hotel Service"]
    E1 --> E2{"Hotel Exists?"}
    E2 -->|No| E3["404 Not Found"]
    E2 -->|Yes| E4["Forward CreateRoom to Room Service"]
    E4 --> E5["Return created Room"]
    
    C -->|POST /reservations| F["ORCHESTRATION"]
    F --> F1["Extract user_id from JWT"]
    F1 --> F2["Fetch Room from Room Service"]
    F2 --> F3{"Room Exists?"}
    F3 -->|No| F4["404 Not Found"]
    F3 -->|Yes| F5["Calculate: nights × room.price"]
    F5 --> F6["Process Payment via Payment Service"]
    F6 --> F7{"Payment Succeeded?"}
    F7 -->|No| F8["Return payment error"]
    F7 -->|Yes| F9["Create Booking via Booking Service"]
    F9 --> F10["Confirm Booking"]
    F10 --> F11["Return Booking JSON"]
    
    C -->|GET /reservations/id/details| G["AGGREGATION"]
    G --> G1["Fetch Booking"]
    G1 --> G2["Fetch Hotel"]
    G2 --> G3["Fetch Room"]
    G3 --> G4["Merge into BookingDetails"]
    G4 --> G5["Return composite JSON"]
    
    C -->|Passthrough| H["Forward to downstream service"]
    H --> H1["Return downstream response"]
```

## Use Case Diagram

```mermaid
graph LR
    subgraph Actors
        User["👤 Authenticated User"]
        Admin["🔑 Admin"]
        Frontend["🖥️ Vue.js Frontend"]
    end
    
    subgraph "BFF Service"
        UC1["View Hotel with Rooms"]
        UC2["Create Room (bridged)"]
        UC3["Create Reservation (orchestrated)"]
        UC4["View Reservation Details"]
        UC5["List User Reservations"]
        UC6["Browse Hotels"]
    end
    
    subgraph "Downstream Services"
        Hotel["🏨 Hotel Service"]
        Room["🛏️ Room Service"]
        Booking["📅 Booking Service"]
        Payment["💳 Payment Service"]
    end
    
    Frontend --> UC1
    Frontend --> UC3
    Frontend --> UC4
    Frontend --> UC5
    Frontend --> UC6
    Admin --> UC2
    
    UC1 --> Hotel
    UC1 --> Room
    UC2 --> Hotel
    UC2 --> Room
    UC3 --> Room
    UC3 --> Payment
    UC3 --> Booking
    UC4 --> Booking
    UC4 --> Hotel
    UC4 --> Room
```

## State Diagram

```mermaid
stateDiagram-v2
    [*] --> Initializing
    Initializing --> Ready : All clients created
    Ready --> Processing : Request received
    Processing --> Ready : Response sent
    Ready --> Degraded : Downstream health check fails
    Degraded --> Ready : Downstream recovers
    Ready --> ShuttingDown : SIGTERM/SIGINT
    Degraded --> ShuttingDown : SIGTERM/SIGINT
    ShuttingDown --> [*] : Graceful shutdown (30s)
    
    state Processing {
        [*] --> Authenticating
        Authenticating --> Routing
        Routing --> Aggregating : Composite endpoint
        Routing --> Bridging : Bridge endpoint
        Routing --> Orchestrating : Orchestration endpoint
        Routing --> Forwarding : Passthrough endpoint
        Aggregating --> [*]
        Bridging --> [*]
        Orchestrating --> [*]
        Forwarding --> [*]
    }
```

## Package Diagram

```mermaid
graph TB
    subgraph "cmd/api"
        Main["main.go"]
    end
    
    subgraph "internal"
        subgraph "handler"
            BaseHandler["handlers.go"]
            HotelH["hotel_handlers.go"]
            RoomH["room_handlers.go"]
            ResH["reservation_handlers.go"]
            MW["middleware.go (JWT, CORS, etc.)"]
            Routes["routing.go"]
        end
        
        subgraph "service"
            SvcIF["service.go (interface + mappings)"]
            HotelSvc["hotel_service.go"]
            RoomSvc["room_service.go"]
            ResSvc["reservation_service.go"]
        end
        
        subgraph "client"
            BaseClient["client.go"]
            HotelClient["hotel_client.go"]
            RoomClient["room_client.go"]
            BookingClient["booking_client.go"]
            ResClient["reservation_client.go"]
            PayClient["payment_client.go"]
            Errors["errors.go"]
        end
        
        subgraph "models"
            Models["models.go"]
        end
        
        subgraph "config"
            Config["config.go"]
        end
    end
    
    Main --> Config
    Main --> Routes
    Main --> SvcIF
    Main --> HotelClient
    Main --> RoomClient
    Main --> BookingClient
    Main --> PayClient
    
    Routes --> BaseHandler
    Routes --> HotelH
    Routes --> RoomH
    Routes --> ResH
    Routes --> MW
    
    HotelH --> SvcIF
    RoomH --> SvcIF
    ResH --> SvcIF
    
    SvcIF --> HotelSvc
    SvcIF --> RoomSvc
    SvcIF --> ResSvc
    
    HotelSvc --> HotelClient
    HotelSvc --> RoomClient
    RoomSvc --> HotelClient
    RoomSvc --> RoomClient
    ResSvc --> BookingClient
    ResSvc --> HotelClient
    ResSvc --> RoomClient
    ResSvc --> PayClient
    
    HotelClient --> BaseClient
    RoomClient --> BaseClient
    BookingClient --> BaseClient
    PayClient --> BaseClient
```

## Reservation Orchestration (Detailed)

The `POST /reservations` endpoint is the most complex flow:

```
1. Extract user_id from JWT claims
2. Validate CreateBookingRequest
3. Fetch Room from Room Service → get price_per_night
4. Parse start_date, end_date
5. Calculate: nights = (end_date - start_date).days
6. Calculate: total_price = nights × price_per_night
7. Process payment via Payment Service
   - Send: booking_id (pre-generated), amount, payment_method_id
   - On failure → return error (no booking created)
8. Create Booking via Booking Service
   - Send: user_id, hotel_id, room_id, dates, total_price, guest info
9. Confirm Booking via Booking Service
   - PATCH status → "confirmed"
10. Return confirmed Booking to frontend
```

## Configuration

```yaml
server:
  port: 8080

downstream_services:
  hotel_service_url: "${HOTEL_SERVICE_URL}"
  room_service_url: "${ROOM_SERVICE_URL}"
  booking_service_url: "${BOOKING_SERVICE_URL}"
  reservation_service_url: "${RESERVATION_SERVICE_URL}"
  payment_service_url: "${PAYMENT_SERVICE_URL}"
  timeout: 30s

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200
```

### Environment Variables

| Variable | Description |
|---|---|
| `HOTEL_SERVICE_URL` | Hotel Service URL (e.g., `http://hotel-service:8080`) |
| `ROOM_SERVICE_URL` | Room Service URL |
| `BOOKING_SERVICE_URL` | Booking Service URL |
| `RESERVATION_SERVICE_URL` | Reservation Service URL (same as booking) |
| `PAYMENT_SERVICE_URL` | Payment Service URL |

### Volume Mounts (Docker)

| Host Path | Container Path | Description |
|---|---|---|
| `./keys/public.pem` | `/app/keys/public.pem` | JWT verification key |

## Port Mapping

| Context | Port |
|---|---|
| Internal (container) | `8080` |
| External (host) | `8087` |
