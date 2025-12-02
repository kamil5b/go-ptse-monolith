package core

import (
	// Shared packages
	"go-modular-monolith/internal/shared/events"
	"go-modular-monolith/internal/shared/uow"

	// Product module
	productDomain "go-modular-monolith/internal/modules/product/domain"
	handlerUnimplemented "go-modular-monolith/internal/modules/product/handler/noop"
	handlerV1 "go-modular-monolith/internal/modules/product/handler/v1"
	repoMongo "go-modular-monolith/internal/modules/product/repository/mongo"
	repoSQL "go-modular-monolith/internal/modules/product/repository/sql"
	serviceUnimplemented "go-modular-monolith/internal/modules/product/service/noop"
	serviceV1 "go-modular-monolith/internal/modules/product/service/v1"

	// User module
	userDomain "go-modular-monolith/internal/modules/user/domain"
	handlerV1User "go-modular-monolith/internal/modules/user/handler/v1"
	repoSQLUser "go-modular-monolith/internal/modules/user/repository/sql"
	serviceV1User "go-modular-monolith/internal/modules/user/service/v1"

	// Auth module
	authACL "go-modular-monolith/internal/modules/auth/acl"
	authDomain "go-modular-monolith/internal/modules/auth/domain"
	handlerNoopAuth "go-modular-monolith/internal/modules/auth/handler/noop"
	handlerV1Auth "go-modular-monolith/internal/modules/auth/handler/v1"
	"go-modular-monolith/internal/modules/auth/middleware"
	repoMongoAuth "go-modular-monolith/internal/modules/auth/repository/mongo"
	repoNoopAuth "go-modular-monolith/internal/modules/auth/repository/noop"
	repoSQLAuth "go-modular-monolith/internal/modules/auth/repository/sql"
	serviceNoopAuth "go-modular-monolith/internal/modules/auth/service/noop"
	serviceV1Auth "go-modular-monolith/internal/modules/auth/service/v1"

	// Unit of Work
	"go-modular-monolith/internal/modules/unitofwork"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	// Event Bus (shared)
	EventBus events.EventBus

	// Product module
	ProductRepository productDomain.Repository
	ProductService    productDomain.Service
	ProductHandler    productDomain.Handler

	// User module
	UserRepository userDomain.Repository
	UserService    userDomain.Service
	UserHandler    userDomain.Handler

	// Auth module
	AuthRepository authDomain.Repository
	AuthService    authDomain.Service
	AuthHandler    authDomain.Handler
	AuthMiddleware *middleware.AuthMiddleware
}

func NewContainer(
	featureFlag FeatureFlag,
	config *Config,
	db *sqlx.DB,
	mongoClient *mongo.Client,
) *Container {
	var (
		productRepository productDomain.Repository
		productService    productDomain.Service
		productHandler    productDomain.Handler
		userRepository    userDomain.Repository
		userService       userDomain.Service
		userHandler       userDomain.Handler
		authRepository    authDomain.Repository
		authService       authDomain.Service
		authHandler       authDomain.Handler
		authMiddleware    *middleware.AuthMiddleware
		unitOfWork        uow.UnitOfWork
	)

	// Initialize event bus (shared across all modules)
	eventBus := events.NewInMemoryEventBus()

	// repo
	switch featureFlag.Repository.Product {
	case "mongo":
		productRepository = repoMongo.NewMongoRepository(mongoClient, config.App.Database.Mongo.MongoDB)
	case "postgres":
		productRepository = repoSQL.NewSQLRepository(db)
	default:
		// productRepository = repoUnimplemented.NewUnimplementedRepository()
	}

	// uow
	unitOfWork = unitofwork.NewDefaultUnitOfWork(db, mongoClient)

	// service
	switch featureFlag.Service.Product {
	case "v1":
		productService = serviceV1.NewServiceV1(productRepository, unitOfWork, eventBus)
	default:
		productService = serviceUnimplemented.NewUnimplementedService()
	}

	// handler
	switch featureFlag.Handler.Product {
	case "v1":
		productHandler = handlerV1.NewHandler(productService)
	default:
		productHandler = handlerUnimplemented.NewUnimplementedHandler()
	}

	// user repo
	switch featureFlag.Repository.User {
	case "postgres":
		userRepository = repoSQLUser.NewSQLRepository(db)
	default:
	}

	// user service
	switch featureFlag.Service.User {
	case "v1":
		userService = serviceV1User.NewServiceV1(userRepository, eventBus)
	default:
	}

	// user handler
	switch featureFlag.Handler.User {
	case "v1":
		userHandler = handlerV1User.NewHandler(userService)
	default:
	}

	// auth repo
	switch featureFlag.Repository.Authentication {
	case "mongo":
		authRepository = repoMongoAuth.NewMongoRepository(mongoClient, "appdb")
	case "postgres":
		authRepository = repoSQLAuth.NewSQLRepository(db)
	default:
		authRepository = repoNoopAuth.NewNoopRepository()
	}

	// auth service
	switch featureFlag.Service.Authentication {
	case "v1":
		authConfig := serviceV1Auth.DefaultAuthConfig()
		if config != nil && config.App.JWT.Secret != "" {
			authConfig.JWTSecret = config.App.JWT.Secret
		}
		// Create ACL adapter for user creation - auth module doesn't directly depend on user module
		userCreator := authACL.NewUserCreatorAdapter(userRepository)
		authService = serviceV1Auth.NewServiceV1(authRepository, userCreator, authConfig)
	default:
		authService = serviceNoopAuth.NewNoopService()
	}

	// auth handler
	switch featureFlag.Handler.Authentication {
	case "v1":
		authHandler = handlerV1Auth.NewHandler(authService)
	default:
		authHandler = handlerNoopAuth.NewNoopHandler()
	}

	// auth middleware
	middlewareConfig := middleware.DefaultMiddlewareConfig()
	if config != nil && config.App.Auth.Type != "" {
		middlewareConfig.AuthType = middleware.AuthType(config.App.Auth.Type)
	}
	if config != nil && config.App.Auth.SessionCookie != "" {
		middlewareConfig.SessionCookie = config.App.Auth.SessionCookie
	}
	authMiddleware = middleware.NewAuthMiddleware(authService, middlewareConfig)

	return &Container{
		EventBus:          eventBus,
		ProductRepository: productRepository,
		ProductService:    productService,
		ProductHandler:    productHandler,
		UserRepository:    userRepository,
		UserService:       userService,
		UserHandler:       userHandler,
		AuthRepository:    authRepository,
		AuthService:       authService,
		AuthHandler:       authHandler,
		AuthMiddleware:    authMiddleware,
	}
}
