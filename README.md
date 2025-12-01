# Automatic Message Sender

An enterprise-grade automatic message sending system built with Go that processes and sends messages from MongoDB at scheduled intervals. The system features a robust worker pattern, Redis caching, comprehensive error handling, and full API documentation.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Database Design](#database-design)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Development](#development)

## ✨ Features

- **Automatic Message Sending**: Fetches and sends unsent messages every 2 minutes (configurable)
- **Worker Pattern**: Robust background worker with thread-safe start/stop controls
- **Redis Caching**: Caches sent message IDs with timestamps (bonus feature)
- **Retry Mechanism**: Automatic retry with exponential backoff for failed operations
- **Status Tracking**: Tracks message status (pending, sent, failed)
- **Character Limit Validation**: Enforces 1000-character limit on message content
- **Prevents Duplicates**: Ensures messages are not sent multiple times
- **RESTful API**: Clean API design with proper HTTP methods
- **Swagger Documentation**: Auto-generated API documentation
- **Docker Support**: Full Docker and Docker Compose configuration
- **Comprehensive Testing**: Unit tests for service, worker, and client layers
- **Graceful Shutdown**: Proper cleanup on application termination
- **Structured Logging**: go-kit based structured logging
- **Clean Architecture**: Follows clean architecture and SOLID principles

## 🏗️ Architecture

The project follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Transport Layer                  │
│  (Handlers, Middleware, Request/Response Encoding)       │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                    Endpoint Layer                        │
│        (Business Logic Entry Points)                     │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                    Service Layer                         │
│  (Business Logic, Worker Management, Orchestration)      │
└────────┬───────────┴───────────┬────────────────────────┘
         │                       │
┌────────▼────────┐    ┌────────▼────────┐
│  Worker Pattern │    │  External APIs  │
│  (Background    │    │  (Message Hook) │
│   Processing)   │    │                 │
└────────┬────────┘    └─────────────────┘
         │
┌────────▼─────────────────────────────────────────────────┐
│                    Data Layer                             │
│         MongoDB Store      │      Redis Cache             │
└──────────────────────────────────────────────────────────┘
```

### Design Patterns Used

- **Repository Pattern**: Abstracts data access logic
- **Worker Pattern**: Background job processing with lifecycle management
- **Dependency Injection**: Loose coupling between components
- **Middleware Pattern**: Cross-cutting concerns (logging, error handling)
- **Strategy Pattern**: Configurable message sending strategies

## 📦 Requirements

- **Go**: 1.22.0 or higher
- **MongoDB**: 4.4 or higher
- **Redis**: 6.0 or higher (optional, for caching)
- **Docker & Docker Compose**: Latest version (optional)

## 🚀 Installation

### Option 1: With Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/mkaykisiz/sender.git
cd sender

# Create .env file
cp .env.example .env

# Edit .env with your configuration
# vim .env

# Start all services
docker-compose up -d

# API available at: http://localhost:8000
# Swagger UI: http://localhost:8000/docs
```

### Option 2: Manual Setup

```bash
# Clone the repository
git clone https://github.com/mkaykisiz/sender.git
cd sender

# Download dependencies
go mod download

# Create .env file
cp .env.example .env

# Edit .env with your configuration
# vim .env

# Start MongoDB (if not running)
mongod --dbpath /path/to/data

# Start Redis (optional)
redis-server

# Run the application
go run cmd/main.go
```

## ⚙️ Configuration

Create a `.env` file in the project root with .env.example:

### Key Configuration Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_START_MESSAGE_COUNT` | Number of messages to process per batch | 2 |
| `CONFIG_SEND_MESSAGE_DURATION` | Interval between message processing | 120s |
| `MESSAGE_CLIENT_URL` | Webhook URL for sending messages | Required |
| `MESSAGE_CLIENT_AUTH_KEY` | Authentication key for webhook | Required |
| `HTTP_SERVER_ADDRESS` | HTTP server listen address | :8000 |

## 🔌 API Endpoints

### Health Check
```http
GET /health
```

### Message Sending Control
```http
POST /start-stop-sending
Content-Type: application/json

{
  "action": "start"  // or "stop"
}
```

**Response:**
```json
{
  "status": "started",  // or "stopped"
  "result": null
}
```

### Retrieve Sent Messages
```http
GET /retrieve-sent-messages
```

**Response:**
```json
{
  "messages": [
    {
      "id": "507f1f77bcf86cd799439011",
      "content": "Message content",
      "recipient": "+905551234567",
      "status": "sent",
      "sent_at": "2024-12-01T00:00:00Z",
      "created_at": "2024-11-30T23:00:00Z"
    }
  ],
  "result": null
}
```

### Swagger Documentation
```http
GET /docs
```

Access the interactive API documentation at `http://localhost:8000/docs`

## 🗄️ Database Design

### MongoDB Collection: `messages`

```javascript
{
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "content": "Message content (max 1000 chars)",
  "recipient": "+905551234567",
  "status": "pending",  // pending | sent | failed
  "sent_at": ISODate("2024-12-01T00:00:00Z"),  // nullable
  "created_at": ISODate("2024-11-30T23:00:00Z")
}
```

**Indexes:**

The following indexes are automatically created when using Docker Compose (via `scripts/init-mongo.js`):

```javascript
// Index 1: For efficient querying of unsent messages (pending/failed)
// Used by the worker to fetch messages that need to be sent
db.messages.createIndex(
  { "status": 1, "created_at": 1 },
  { 
    name: "idx_status_created_at",
    background: true 
  }
);

// Index 2: For retrieving sent messages
// Used by the retrieve-sent-messages API endpoint
db.messages.createIndex(
  { "status": 1 },
  { 
    name: "idx_status",
    background: true 
  }
);
```

**Manual Index Creation:**

If not using Docker Compose, create indexes manually:

```bash
# Connect to MongoDB
mongosh mongodb://localhost:27017/sender_db

# Run the initialization script
load('scripts/init-mongo.js')

# Or create indexes manually
db.messages.createIndex({ "status": 1, "created_at": 1 }, { name: "idx_status_created_at", background: true });
db.messages.createIndex({ "status": 1 }, { name: "idx_status", background: true });
```

### Redis Cache Structure

```
Key: "message:{messageId}"
Value: "{timestamp}"
TTL: No expiration (persistent cache)
```

**Purpose**: Prevents duplicate message sending and provides quick lookup for sent messages.

## 🧪 Testing

### Run All Tests
```bash
go test ./... -v
```

### Run Specific Package Tests
```bash
# Service tests
go test ./internal/service/... -v

# Worker tests
go test ./internal/service/... -v -run TestWorker

# Message client tests
go test ./internal/client/messagehook/... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

- **Service Layer**: Core business logic and worker management
- **Worker**: Background processing, message fetching, and sending
- **Message Client**: HTTP client for webhook integration
- **Mocks**: Complete mock implementations for all external dependencies

## 📁 Project Structure

```
.
├── cmd/
│   └── main.go                 # Application entry point
├── configs/
│   └── env-vars/              # Environment variable configuration
├── docs/
│   ├── docs.go                # Swagger documentation
│   └── swagger.yaml           # Generated Swagger spec
├── scripts/
│   └── init-mongo.js          # MongoDB initialization script
├── internal/
│   ├── client/
│   │   └── messagehook/       # Message webhook client
│   ├── endpoints/             # Go-kit endpoints
│   ├── localization/          # i18n support
│   ├── middlewares/           # Logging and other middlewares
│   ├── mock/                  # Mock implementations for testing
│   │   ├── client/
│   │   └── store/
│   ├── service/               # Business logic and worker
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── worker.go
│   │   └── worker_test.go
│   ├── store/
│   │   ├── mongo/             # MongoDB implementation
│   │   └── redis/             # Redis implementation
│   └── transport/
│       └── http/              # HTTP handlers and routing
├── sender.go                  # Core domain types
├── docker-compose.yaml        # Docker Compose configuration
├── Dockerfile                 # Application Dockerfile
├── Makefile                   # Build and utility commands
├── go.mod                     # Go module definition
├── .env.example               # Example environment configuration
└── README.md                  # This file
```

## 🛠️ Development

### Available Make Commands

```bash
# Generate Swagger documentation
make swagger

# Serve Swagger UI
make serve-swagger

# Run tests
make test

# Generate test coverage
make cover
```

### Code Quality Standards

- **Go fmt**: Code is formatted with `gofmt`
- **Linting**: Follow standard Go linting rules
- **Error Handling**: Comprehensive error handling with structured logging
- **Testing**: Unit tests for critical components
- **Documentation**: Inline comments and Swagger documentation

### Adding New Features

1. **Define Domain Types**: Add new types to `sender.go`
2. **Implement Service Logic**: Add methods to `internal/service/service.go`
3. **Create Endpoints**: Add endpoints to `internal/endpoints/endpoints.go`
4. **Add HTTP Handlers**: Update `internal/transport/http/http.go`
5. **Update Swagger**: Add documentation to `docs/docs.go`
6. **Write Tests**: Add unit tests for new functionality
7. **Generate Swagger**: Run `make swagger`

## 📝 Environment Variables Reference

### Service Configuration
- `SERVICE_PROJECT_NAME`: Project identifier
- `SERVICE_NAME`: Service name
- `SERVICE_ENVIRONMENT`: Environment (local/dev/prod)
- `SERVICE_SHUTDOWN_SLEEP_DURATION`: Graceful shutdown delay

### HTTP Server
- `HTTP_SERVER_ADDRESS`: Server listen address
- `HTTP_SERVER_READ_TIMEOUT`: Request read timeout
- `HTTP_SERVER_WRITE_TIMEOUT`: Response write timeout
- `HTTP_SERVER_SHUTDOWN_TIMEOUT`: Graceful shutdown timeout

### MongoDB
- `MONGO_URI`: MongoDB connection string
- `MONGO_DATABASE`: Database name
- `MONGO_CONNECT_TIMEOUT`: Connection timeout
- `MONGO_READ_TIMEOUT`: Read operation timeout
- `MONGO_WRITE_TIMEOUT`: Write operation timeout

### Redis
- `REDIS_ADDRESS`: Redis server address
- `REDIS_PASSWORD`: Redis password (optional)
- `REDIS_DB`: Redis database number
- `REDIS_POOL_SIZE`: Connection pool size

### Message Client
- `MESSAGE_CLIENT_URL`: Webhook endpoint URL
- `MESSAGE_CLIENT_AUTH_KEY`: Authentication key
- `MESSAGE_CLIENT_TIMEOUT`: Request timeout
- `MESSAGE_CLIENT_MAX_RETRIES`: Maximum retry attempts
- `MESSAGE_CLIENT_RETRY_DELAY`: Delay between retries

### Worker Configuration
- `CONFIG_START_MESSAGE_COUNT`: Messages per batch
- `CONFIG_SEND_MESSAGE_DURATION`: Processing interval

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 👥 Authors

- **Mehmet Kaykısız** - [GitHub](https://github.com/mkaykisiz)

## 🙏 Acknowledgments

- Go-kit for the excellent microservice toolkit
- MongoDB for the robust database
- Redis for high-performance caching
- The Go community for amazing libraries and tools