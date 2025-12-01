package core

import (
	"go-modular-monolith/internal/domain/auth"
	"go-modular-monolith/internal/domain/product"
	"go-modular-monolith/internal/domain/uow"
	"go-modular-monolith/internal/domain/user"

	repoMongo "go-modular-monolith/internal/modules/product/repository/mongo"
	repoSQL "go-modular-monolith/internal/modules/product/repository/sql"
	"go-modular-monolith/internal/modules/unitofwork"

	serviceUnimplemented "go-modular-monolith/internal/modules/product/service/noop"
	serviceV1 "go-modular-monolith/internal/modules/product/service/v1"

	handlerUnimplemented "go-modular-monolith/internal/modules/product/handler/noop"
	handlerV1 "go-modular-monolith/internal/modules/product/handler/v1"

	handlerV1User "go-modular-monolith/internal/modules/user/handler/v1"
	repoSQLUser "go-modular-monolith/internal/modules/user/repository/sql"
	serviceV1User "go-modular-monolith/internal/modules/user/service/v1"

	// Auth imports
	handlerNoopAuth "go-modular-monolith/internal/modules/auth/handler/noop"
	handlerV1Auth "go-modular-monolith/internal/modules/auth/handler/v1"
	"go-modular-monolith/internal/modules/auth/middleware"
	repoMongoAuth "go-modular-monolith/internal/modules/auth/repository/mongo"
	repoNoopAuth "go-modular-monolith/internal/modules/auth/repository/noop"
	repoSQLAuth "go-modular-monolith/internal/modules/auth/repository/sql"
	serviceNoopAuth "go-modular-monolith/internal/modules/auth/service/noop"
	serviceV1Auth "go-modular-monolith/internal/modules/auth/service/v1"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	ProductRepository product.ProductRepository
	ProductService    product.ProductService
	ProductHandler    product.ProductHandler
	UserRepository    user.UserRepository
	UserService       user.UserService
	UserHandler       user.UserHandler
	AuthRepository    auth.AuthRepository
	AuthService       auth.AuthService
	AuthHandler       auth.AuthHandler
	AuthMiddleware    *middleware.AuthMiddleware
}

func NewContainer(
	featureFlag FeatureFlag,
	config *Config,
	db *sqlx.DB,
	mongoClient *mongo.Client,
) *Container {
	var (
		productRepository product.ProductRepository
		productService    product.ProductService
		productHandler    product.ProductHandler
		userRepository    user.UserRepository
		userService       user.UserService
		userHandler       user.UserHandler
		authRepository    auth.AuthRepository
		authService       auth.AuthService
		authHandler       auth.AuthHandler
		authMiddleware    *middleware.AuthMiddleware
		unitOfWork        uow.UnitOfWork
	)

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
		productService = serviceV1.NewServiceV1(productRepository, unitOfWork)
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
		userService = serviceV1User.NewServiceV1(userRepository)
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
		authService = serviceV1Auth.NewServiceV1(authRepository, userRepository, authConfig)
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
		ProductService: productService,
		ProductHandler: productHandler,
		UserRepository: userRepository,
		UserService:    userService,
		UserHandler:    userHandler,
		AuthRepository: authRepository,
		AuthService:    authService,
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
	}
}
