#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# New Module Scaffold Script
# ============================================================================
# Scaffolds a new module with the correct structure and boilerplate
# Usage: ./scripts/new-module.sh <module_name>
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Validate arguments
if [ $# -ne 1 ]; then
  echo "${RED}❌ Usage: ./scripts/new-module.sh <module_name>${NC}"
  echo ""
  echo "Example: ./scripts/new-module.sh order"
  exit 1
fi

MODULE_NAME="$1"

# Validate module name (lowercase alphanumeric only)
if ! [[ "$MODULE_NAME" =~ ^[a-z][a-z0-9]*$ ]]; then
  echo "${RED}❌ Invalid module name: $MODULE_NAME${NC}"
  echo "   Module names must be lowercase and start with a letter"
  exit 1
fi

MODULE_PATH="internal/modules/$MODULE_NAME"

# Check if module already exists
if [ -d "$MODULE_PATH" ]; then
  echo "${RED}❌ Module already exists: $MODULE_PATH${NC}"
  exit 1
fi

echo "${BLUE}📦 Creating new module: $MODULE_NAME${NC}"
echo ""

# Create directory structure
echo "${BLUE}📁 Creating directories...${NC}"
mkdir -p "$MODULE_PATH/domain"
mkdir -p "$MODULE_PATH/handler/v1"
mkdir -p "$MODULE_PATH/handler/noop"
mkdir -p "$MODULE_PATH/service/v1"
mkdir -p "$MODULE_PATH/service/noop"
mkdir -p "$MODULE_PATH/repository/sql"
mkdir -p "$MODULE_PATH/repository/mongo"
mkdir -p "$MODULE_PATH/repository/noop"
mkdir -p "$MODULE_PATH/acl"
mkdir -p "$MODULE_PATH/domain/mocks"

echo "   ✅ Directories created"
echo ""

# Create domain files
echo "${BLUE}📝 Creating domain layer...${NC}"

# model.go
cat > "$MODULE_PATH/domain/model.go" << 'EOF'
package domain

// TODO: Define your domain models here
// Example:
// type Product struct {
//     ID    string
//     Name  string
//     Price float64
// }
EOF

# interfaces.go
cat > "$MODULE_PATH/domain/interfaces.go" << 'EOF'
package domain

import sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"

// TODO: Define your module's interfaces here

// Handler interface
type Handler interface {
    // Create(c sharedctx.Context) error
    // Get(c sharedctx.Context) error
    // List(c sharedctx.Context) error
    // Update(c sharedctx.Context) error
    // Delete(c sharedctx.Context) error
}

// Service interface
type Service interface {
    // Define your service methods here
}

// Repository interface
type Repository interface {
    // Define your repository methods here
}
EOF

# request.go
cat > "$MODULE_PATH/domain/request.go" << 'EOF'
package domain

// TODO: Define your request DTOs here
// Example:
// type CreateProductRequest struct {
//     Name  string `json:"name" validate:"required"`
//     Price float64 `json:"price" validate:"required,gt=0"`
// }
EOF

# response.go
cat > "$MODULE_PATH/domain/response.go" << 'EOF'
package domain

// TODO: Define your response DTOs here
// Example:
// type ProductResponse struct {
//     ID    string  `json:"id"`
//     Name  string  `json:"name"`
//     Price float64 `json:"price"`
// }
EOF

# events.go
cat > "$MODULE_PATH/domain/events.go" << 'EOF'
package domain

import "time"

// TODO: Define your domain events here
// Example:
// type ProductCreated struct {
//     ProductID string
//     Name      string
//     Timestamp time.Time
// }
//
// func (e ProductCreated) EventName() string    { return "product.created" }
// func (e ProductCreated) OccurredAt() time.Time { return e.Timestamp }
EOF

echo "   ✅ Domain layer created"
echo ""

# Create handler files
echo "${BLUE}📝 Creating handler layer...${NC}"

cat > "$MODULE_PATH/handler/v1/handler_v1.$MODULE_NAME.go" << 'EOF'
package v1

import (
	"context"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/MODULEPLACEHOLDER/domain"
	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

type Handler struct {
	service domain.Service
}

func NewHandler(service domain.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// TODO: Implement your handler methods here
// Example:
// func (h *Handler) Create(c sharedctx.Context) error {
//     var req domain.CreateProductRequest
//     if err := c.BindJSON(&req); err != nil {
//         return err
//     }
//     // Call service...
//     return c.JSON(201, result)
// }
EOF

# Replace placeholder
sed -i '' "s/MODULEPLACEHOLDER/$MODULE_NAME/g" "$MODULE_PATH/handler/v1/handler_v1.$MODULE_NAME.go"

cat > "$MODULE_PATH/handler/noop/handler_noop.$MODULE_NAME.go" << 'EOF'
package noop

import sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"

type NoOpHandler struct{}

func NewNoOpHandler() *NoOpHandler {
	return &NoOpHandler{}
}

// TODO: Implement no-op versions of your handler methods
EOF

echo "   ✅ Handler layer created"
echo ""

# Create service files
echo "${BLUE}📝 Creating service layer...${NC}"

cat > "$MODULE_PATH/service/v1/service_v1.$MODULE_NAME.go" << 'EOF'
package v1

import (
	"context"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/MODULEPLACEHOLDER/domain"
)

type Service struct {
	repository domain.Repository
}

func NewService(repository domain.Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// TODO: Implement your service methods here
EOF

sed -i '' "s/MODULEPLACEHOLDER/$MODULE_NAME/g" "$MODULE_PATH/service/v1/service_v1.$MODULE_NAME.go"

cat > "$MODULE_PATH/service/noop/service_noop.$MODULE_NAME.go" << 'EOF'
package noop

type NoOpService struct{}

func NewNoOpService() *NoOpService {
	return &NoOpService{}
}

// TODO: Implement no-op versions of your service methods
EOF

echo "   ✅ Service layer created"
echo ""

# Create repository files
echo "${BLUE}📝 Creating repository layer...${NC}"

cat > "$MODULE_PATH/repository/sql/repository_sql.$MODULE_NAME.go" << 'EOF'
package sql

import (
	"context"
	"database/sql"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/MODULEPLACEHOLDER/domain"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

// TODO: Implement your SQL repository methods here
EOF

sed -i '' "s/MODULEPLACEHOLDER/$MODULE_NAME/g" "$MODULE_PATH/repository/sql/repository_sql.$MODULE_NAME.go"

cat > "$MODULE_PATH/repository/mongo/repository_mongo.$MODULE_NAME.go" << 'EOF'
package mongo

import (
	"context"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/MODULEPLACEHOLDER/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	client *mongo.Client
	db     string
}

func NewMongoRepository(client *mongo.Client, db string) *MongoRepository {
	return &MongoRepository{
		client: client,
		db:     db,
	}
}

// TODO: Implement your MongoDB repository methods here
EOF

sed -i '' "s/MODULEPLACEHOLDER/$MODULE_NAME/g" "$MODULE_PATH/repository/mongo/repository_mongo.$MODULE_NAME.go"

cat > "$MODULE_PATH/repository/noop/repository_noop.$MODULE_NAME.go" << 'EOF'
package noop

type NoOpRepository struct{}

func NewNoOpRepository() *NoOpRepository {
	return &NoOpRepository{}
}

// TODO: Implement no-op versions of your repository methods
EOF

echo "   ✅ Repository layer created"
echo ""

# Create ACL stub
echo "${BLUE}📝 Creating ACL placeholder...${NC}"

cat > "$MODULE_PATH/acl/.gitkeep" << 'EOF'
# Anti-Corruption Layer (ACL)
# 
# Create ACL adapters here when this module needs to communicate
# with other modules. ACL adapters allow cross-module imports.
#
# Example: user_creator.go
# 
# This folder is allowed to import from other modules.
# See TECHNICAL_DOCUMENTATION.md for ACL pattern details.
EOF

echo "   ✅ ACL placeholder created"
echo ""

# Create .gitkeep files to preserve directory structure
touch "$MODULE_PATH/domain/mocks/.gitkeep"

echo "${GREEN}✅ Module created successfully!${NC}"
echo ""
echo "${BLUE}Next steps:${NC}"
echo "  1. Define your domain models in: $MODULE_PATH/domain/model.go"
echo "  2. Define your interfaces in: $MODULE_PATH/domain/interfaces.go"
echo "  3. Implement handlers in: $MODULE_PATH/handler/v1/"
echo "  4. Implement services in: $MODULE_PATH/service/v1/"
echo "  5. Implement repositories in: $MODULE_PATH/repository/sql/ and mongo/"
echo "  6. Wire up in internal/app/core/container.go"
echo "  7. Add routes in internal/app/http/routes.go"
echo "  8. Add feature flags in config/featureflags.yaml"
echo ""
echo "${YELLOW}Verify module isolation:${NC}"
echo "  ./scripts/lint-deps-check.sh --verbose"
echo ""
