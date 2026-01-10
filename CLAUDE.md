# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Local development (from src/ directory)
cd src
go run main.go

# Build binary
cd src
go build -o orderease.exe main.go

# Run tests
cd src
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate Wire dependency injection code
cd src/application/services
go generate
# or: wire generate
```

## Docker Deployment

```bash
# All-in-one build (Go backend + Frontend UIs + Backend UI)
docker build -f Dockerfile.all-in-one -t orderease:latest .

# Run with docker-compose
docker-compose -f docker-compose.all-in-one.yml up -d --build
```

## Architecture Overview

OrderEase is a **DDD (Domain-Driven Design) four-layer architecture** monolithic application:

```
┌─────────────────────────────────────────┐
│   Interfaces Layer (接口层)              │  HTTP handlers, middleware, routes
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│  Application Layer (应用层)              │  Services orchestrate use cases
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│    Domain Layer (领域层)                 │  Core business logic (entities, value objects)
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ Infrastructure Layer (基础设施层)        │  Repository implementations, persistence
└─────────────────────────────────────────┘
```

## Directory Structure

```
src/
├── domain/                    # Domain Layer - Core business logic
│   ├── order/                 # Order bounded context
│   ├── product/               # Product bounded context
│   ├── shop/                  # Shop bounded context
│   ├── user/                  # User bounded context
│   └── shared/                # Shared kernel (ID, Price, constants)
│
├── application/              # Application Layer - Use case orchestration
│   ├── services/             # Application services (use Wire DI)
│   └── dto/                  # Data transfer objects
│
├── infrastructure/           # Infrastructure Layer - Technical details
│   ├── persistence/          # Domain ↔ DB model converters
│   └── repositories/        # Repository implementations (GORM)
│
├── interfaces/              # Interfaces Layer - HTTP handlers
│   ├── http/
│   └── middleware/
│
├── routes/                  # Route definitions (backend/, frontend/)
├── models/                  # GORM database models
├── handlers/                # Legacy handlers (compatibility layer)
└── main.go                  # Application entry point
```

## Dependency Injection (Wire Framework)

**Wire config**: `src/application/services/wire.go`

After modifying Wire providers, regenerate the code:
```bash
cd src/application/services
go generate
```

**Service Container Usage**:
```go
container, err := services.InitializeServiceContainer(db)
orderService := container.OrderService
```

## Domain Layer Conventions

### Shared Kernel

Always use shared types for IDs and prices:
```go
type Order struct {
    ID         shared.ID    // ✅ Correct
    TotalPrice shared.Price  // ✅ Correct
}
```

### Repository Pattern

- **Interface** defined in `domain/*/repository.go`
- **Implementation** in `infrastructure/repositories/*_repository.go`

```go
// domain/order/repository.go
type OrderRepository interface {
    Save(order *Order) error
    FindByID(id shared.ID) (*Order, error)
}

// infrastructure/repositories/order_repository.go
type OrderRepositoryImpl struct { db *gorm.DB }
```

### Business Logic Location

- **Domain entities** contain business rules (validation, state transitions)
- **Domain services** handle cross-aggregate logic
- **Application services** orchestrate only (transaction management)

```go
// Domain: business logic
func (o *Order) ValidateItems(finder ProductFinder) error
func (o *Order) CalculateTotal(finder ProductFinder) error

// Application: orchestration
func (s *OrderService) CreateOrder(req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    ord.ValidateItems(finder)
    ord.CalculateTotal(finder)
    return s.executeCreateOrderTransaction(ord, finder)
}
```

## Bounded Contexts

The project has four bounded contexts:

1. **User Context** (`domain/user/`) - Authentication, permissions
2. **Shop Context** (`domain/shop/`) - Shop management, tags
3. **Product Context** (`domain/product/`) - Product catalog, inventory
4. **Order Context** (`domain/order/`) - Order lifecycle, status flows

## Constants

All domain constants are centralized in `domain/shared/constants.go`:
- `OrderStatus*` - Order status constants
- `ProductStatus*` - Product status constants

## Transaction Management

Use the transaction template in `application/services/transaction.go`:
```go
err = WithTx(s.db, func(tx *gorm.DB) error {
    // Operations within transaction
    return nil
})
```

## Configuration

Configuration file: `src/config/config.yaml`

Key settings:
- `server.port`: 8080
- `database.*`: MySQL connection
- `jwt.secret`: JWT signing key

## Important Design Documents

- `DDD_战略设计方案.md` - DDD strategic design (3000+ lines)
- `src/application/services/WIRE_USAGE.md` - Wire usage guide
- `.trae/documents/业务逻辑分析报告.md` - Business logic analysis

## Naming Conventions

- **Repositories**: `XxxRepository` interface, `xxxRepository` implementation
- **Services**: `XxxService` struct, `NewXxxService` constructor
- **Methods**: `CreateXxx`, `UpdateXxx`, `GetXxx`, `ListXxx`, `DeleteXxx`

## Legacy Code

The `src/handlers/` directory contains old handler code that bypasses the DDD layers. New code should use `interfaces/http/` and application services instead.
