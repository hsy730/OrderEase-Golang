# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OrderEase-Golang is a full-stack e-commerce order management system backend built with Go, implementing Domain-Driven Design (DDD) architecture with 98-99% maturity after 70+ refactoring steps.

**Technology Stack:**
- Go 1.21, Gin web framework, GORM ORM, MySQL 8.0
- JWT authentication, Snowflake ID generation
- Cron-based scheduled tasks, Zap structured logging

## Quick Start Commands

### Local Development
```bash
cd src
go run main.go                    # Start server on :8080
go mod tidy                        # Update dependencies
go build -o ../build/OrderEase.exe .  # Build executable
```

### Configuration
Edit `src/config/config.yaml`:
```yaml
server:
  port: 8080
  basePath: "/api/order-ease/v1"

database:
  host: 127.0.0.1
  port: 3306
  username: root
  password: 123456
  dbname: mysql

jwt:
  secret: "e6jf493kdhbms9ew6mv2v1a4dx2"
  expiration: 7200  # 2 hours
```

### Testing
```bash
cd src
go test ./utils -v               # Run unit tests

# Integration tests (Python pytest)
cd ../OrderEase-Deploy/test
pytest -v                          # Run all tests
pytest admin/test_business_flow.py # Run specific suite
pytest -v --html=report.html       # Generate HTML report
```

### Docker Deployment
```bash
docker build -t orderease:latest .
docker-compose up -d               # Standard deployment
docker-compose -f docker-compose.all-in-one.yml up -d  # All-in-one
```

### Stopping the Service
```bash
# Find and kill the process
tasklist | grep -i "main.exe"
taskkill //F //PID <PID>
```

## DDD Architecture

This project follows strict DDD four-layer architecture:

```
Interface Layer (handlers/routes/middleware)
         ↓
Application Layer (domain services)
         ↓
Domain Layer (entities/value objects) ★ Core
         ↓
Infrastructure Layer (repositories/models)
```

### Directory Structure

```
src/
├── domain/           # Domain layer - core business logic
│   ├── order/        # Order aggregate (order.go, service.go, repository.go)
│   ├── shop/         # Shop aggregate
│   ├── product/      # Product aggregate
│   ├── user/         # User aggregate
│   ├── tag/          # Tag entity
│   └── shared/       # Shared components
│       └── value_objects/  # Value objects (Phone, Password, OrderStatus)
├── handlers/         # HTTP handlers (interface layer)
├── repositories/     # Repository implementations (infrastructure)
├── models/           # GORM persistence models
├── routes/           # Route definitions (backend/, frontend/)
├── middleware/       # Auth, CORS, logging middleware
├── config/           # Configuration management
├── utils/            # JWT, Snowflake ID, logging utilities
├── tasks/            # Scheduled tasks (cron)
└── main.go           # Entry point
```

## Key Design Patterns

### 1. Repository Pattern

**Interface** (in domain layer):
```go
// domain/order/repository.go
type Repository interface {
    Create(order *models.Order) error
    GetByID(id snowflake.ID) (*models.Order, error)
    // ...
}
```

**Implementation** (in infrastructure layer):
```go
// repositories/order_repository.go
type OrderRepository struct {
    DB *gorm.DB
}

func (r *OrderRepository) GetByID(id snowflake.ID) (*models.Order, error) {
    // Implementation
}
```

### 2. Factory Pattern for Domain Entities

```go
// domain/order/order.go
func NewOrder(userID snowflake.ID, shopID uint64) *Order {
    return &Order{
        userID:    userID,
        shopID:    shopID,
        status:    value_objects.OrderStatusPending,
        createdAt: time.Now(),
    }
}

// Convert from persistence model
func OrderFromModel(model *models.Order) *Order {
    // ...
}

// Convert to persistence model
func (o *Order) ToModel() *models.Order {
    // ...
}
```

### 3. Value Objects

Located in `domain/shared/value_objects/`:
- `Phone` - 11-digit phone validation
- `Password` - 6-20 chars, letters + digits
- `SimplePassword` - 6 chars for frontend users
- `OrderStatus` - Configurable order status flow

### 4. Rich Domain Model

Business logic is encapsulated in entities:
```go
// domain/order/order.go
func (o *Order) ValidateItems() error
func (o *Order) CalculateTotal() models.Price
func (o *Order) CanTransitionTo(to value_objects.OrderStatus) bool
func (o *Order) CanBeDeleted() bool
```

## API Route Structure

**Base Path**: `/api/order-ease/v1`

### Backend Routes
- **Admin**: `/admin/*` - Admin endpoints (requires admin role)
- **Shop Owner**: `/shopOwner/*` - Shop owner endpoints (requires shop owner auth)
- **No Auth**: `/login`, `/user/login` - Public login endpoints

### Frontend Routes
- **Protected**: `/product/*`, `/order/*`, `/tag/*`, `/shop/*` - Require frontend user auth
- **Public**: `/user/login`, `/user/register` - User registration/login

### Authentication Middleware

```go
// BackendAuthMiddleware(isAdmin bool) - For admin/shop owner routes
middleware.BackendAuthMiddleware(false)  // Shop owner
middleware.BackendAuthMiddleware(true)   // Admin

// FrontendAuthMiddleware() - For frontend user routes
middleware.FrontendAuthMiddleware()
```

## ID Generation and JSON Serialization

### Snowflake ID
```go
// Generate new ID
productModel.ID = utils.GenerateSnowflakeID()

// Type is snowflake.ID (int64)
type Product struct {
    ID snowflake.ID `gorm:"primarykey" json:"id"`
}
```

### JSON Response Format
Go structs return PascalCase by default due to GORM defaults. Always check for both `"ID"` and `"id"` when parsing API responses.

## Default Credentials

- **Admin**: `admin` / `Admin@123456`
- **Default Shop Owner**: Created dynamically during testing

## Domain Aggregates

### Order (`domain/order/`)
- **Entity**: Order, OrderItem
- **Service**: CreateOrder, UpdateOrder, ValidateOrder, RestoreStock
- **Repository**: CRUD, GetByShopID, GetByUserID

### Shop (`domain/shop/`)
- **Entity**: Shop
- **Service**: ValidateForDeletion
- **Repository**: CRUD, GetByName

### Product (`domain/product/`)
- **Entity**: Product
- **Service**: ValidateForDeletion, CanTransitionTo (status flow)
- **Repository**: CRUD, GetProductsByShop with status filtering

### User (`domain/user/`)
- **Entity**: User
- **Service**: Register, UpdatePhone, UpdatePassword
- **Repository**: CRUD, GetByName

## Working with Status Filtering

**Recent Change**: Client queries now only return `online` products, while admin queries return all statuses.

The filtering is done by checking the request path in handlers:
```go
// handlers/product.go
onlyOnline := !strings.HasPrefix(c.Request.URL.Path, "/api/order-ease/v1/shopOwner/") &&
              !strings.HasPrefix(c.Request.URL.Path, "/api/order-ease/v1/admin/")
```

## Important Implementation Notes

### Adding New Features
1. Define domain model in `domain/{aggregate}/`
2. Create GORM model in `models/`
3. Implement repository in `repositories/`
4. Create handler in `handlers/`
5. Register route in `routes/`

### Modifying Existing Features
1. Update Domain layer business logic first
2. Update Repository data access
3. Adjust Handler interface
4. Run tests to verify

### Adding New Aggregates
1. Create `domain/{aggregate}/` directory
2. Implement entities, value objects, domain service
3. Create Repository interface in domain
4. Implement in `repositories/`
5. Add handler in `handlers/`
6. Register route in `routes/`

## Common Pitfalls

1. **JSON Field Names**: GORM returns PascalCase, not camelCase. Always check for `"ID"` vs `"id"`
2. **ID Type**: Use `snowflake.ID` (int64), not `uint64` for new entities
3. **Status Filtering**: Client queries filter to `online` only, admin queries see all statuses
4. **Repository Parameters**: Many repository methods now have `onlyOnline` parameter
5. **Domain Model Conversion**: Always use `FromModel` and `ToModel` methods for conversions
6. **Auth Middleware**: Different middleware for backend vs frontend routes

## Test Execution Order (Priority)

1. Frontend tests (Priority 0)
2. Shop owner business flow (Priority 10) - 25 tests
3. Admin business flow (Priority 20)
4. Auth tests (Priority 100) - logs out tokens
5. Unauthorized tests (Priority 110)

## Related Documentation

- `DDD_ARCHITECTURE.md` - Detailed DDD architecture documentation
- `DDD_REFACTORING_SUMMARY.md` - Complete refactoring history (70 steps)
- `README_DOCKER.md` - Docker deployment guide
- `CLAUDE.md` (parent repo) - Full-stack project overview including Vue frontends
