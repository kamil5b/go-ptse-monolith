package core

import (
	// Shared packages
	"context"

	"github.com/kamil5b/go-ptse-monolith/internal/shared/cache"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/email"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/events"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/uow"
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

	// Worker infrastructure
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	asynqworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/asynq"
	rabbitmqworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/rabbitmq"
	redpandaworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"

	// Email infrastructure
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/email/mailgun"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/email/smtp"

	// Storage infrastructure
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/gcs"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/local"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/noop"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/s3"

	// Product module
	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	handlerUnimplemented "github.com/kamil5b/go-ptse-monolith/internal/modules/product/handler/noop"
	handlerV1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/handler/v1"
	repoMongo "github.com/kamil5b/go-ptse-monolith/internal/modules/product/repository/mongo"
	repoSQL "github.com/kamil5b/go-ptse-monolith/internal/modules/product/repository/sql"
	serviceUnimplemented "github.com/kamil5b/go-ptse-monolith/internal/modules/product/service/noop"
	serviceV1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/service/v1"

	// User module
	userDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
	handlerV1User "github.com/kamil5b/go-ptse-monolith/internal/modules/user/handler/v1"
	repoSQLUser "github.com/kamil5b/go-ptse-monolith/internal/modules/user/repository/sql"
	serviceV1User "github.com/kamil5b/go-ptse-monolith/internal/modules/user/service/v1"

	// Auth module
	authACL "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/acl"
	authDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/domain"
	handlerNoopAuth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/handler/noop"
	handlerV1Auth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/handler/v1"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/auth/middleware"
	repoMongoAuth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/repository/mongo"
	repoNoopAuth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/repository/noop"
	repoSQLAuth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/repository/sql"
	serviceNoopAuth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/service/noop"
	serviceV1Auth "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/service/v1"

	// Unit of Work
	"github.com/kamil5b/go-ptse-monolith/internal/modules/unitofwork"

	// Infrastructure
	infracache "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/cache"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	// Cache (shared)
	Cache cache.Cache

	// Event Bus (shared)
	EventBus events.EventBus

	// Email Service (shared)
	EmailClient email.EmailService

	// Storage Service (shared)
	StorageService storage.StorageService

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

	// Worker (infrastructure)
	WorkerClient sharedworker.Client
	WorkerServer sharedworker.Server
}

func NewContainer(
	featureFlag FeatureFlag,
	config *Config,
	db *sqlx.DB,
	mongoClient *mongo.Client,
) *Container {
	var (
		cacheInstance     cache.Cache
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

	// Initialize cache (shared across all modules)
	switch featureFlag.Cache {
	case "redis":
		if config != nil {
			redisConfig := infracache.RedisConfig{
				Host:         config.App.Redis.Host,
				Port:         config.App.Redis.Port,
				Password:     config.App.Redis.Password,
				DB:           config.App.Redis.DB,
				MaxRetries:   config.App.Redis.MaxRetries,
				PoolSize:     config.App.Redis.PoolSize,
				MinIdleConns: config.App.Redis.MinIdleConns,
			}
			if redisClient, err := infracache.NewRedisClient(redisConfig); err == nil {
				cacheInstance = infracache.NewRedisCache(redisClient)
			} else {
				// Fallback to in-memory cache if Redis connection fails
				cacheInstance = cache.NewInMemoryCache()
			}
		} else {
			cacheInstance = cache.NewInMemoryCache()
		}
	case "memory", "disable":
		fallthrough
	default:
		cacheInstance = cache.NewInMemoryCache()
	}

	// Initialize event bus (shared across all modules)
	eventBus := events.NewInMemoryEventBus()

	// Initialize email service (before modules that depend on it)
	var emailService email.EmailService
	if featureFlag.Email.Enabled && featureFlag.Email.Provider != "noop" && config != nil {
		switch featureFlag.Email.Provider {
		case "smtp":
			emailService = smtp.NewSMTPEmailService(smtp.SMTPConfig{
				Host:     config.App.Email.SMTP.Host,
				Port:     config.App.Email.SMTP.Port,
				Username: config.App.Email.SMTP.Username,
				Password: config.App.Email.SMTP.Password,
				FromAddr: config.App.Email.SMTP.FromAddr,
				FromName: config.App.Email.SMTP.FromName,
			})
		case "mailgun":
			emailService = mailgun.NewMailgunEmailService(mailgun.MailgunConfig{
				Domain:   config.App.Email.Mailgun.Domain,
				APIKey:   config.App.Email.Mailgun.APIKey,
				FromAddr: config.App.Email.Mailgun.FromAddr,
				FromName: config.App.Email.Mailgun.FromName,
			})
		default:
			emailService = email.NewNoOpEmailService()
		}
	} else {
		// Use no-op implementation when email is disabled or provider is noop
		emailService = email.NewNoOpEmailService()
	}

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
		productService = serviceV1.NewServiceV1(productRepository, unitOfWork, eventBus, cacheInstance)
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
		userService = serviceV1User.NewServiceV1(userRepository, eventBus, emailService, cacheInstance)
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

	// Initialize worker client and server
	var workerClient sharedworker.Client
	var workerServer sharedworker.Server

	if featureFlag.Worker.Enabled && featureFlag.Worker.Backend != "disable" && config != nil {
		// Initialize worker client
		switch featureFlag.Worker.Backend {
		case "asynq":
			workerClient = asynqworker.NewAsynqClient(config.App.Worker.Asynq.RedisURL)
		case "rabbitmq":
			if client, err := rabbitmqworker.NewRabbitMQClient(
				config.App.Worker.RabbitMQ.URL,
				config.App.Worker.RabbitMQ.Exchange,
				config.App.Worker.RabbitMQ.Queue,
			); err == nil {
				workerClient = client
			} else {
				workerClient = infraworker.NewNoOpClient()
			}
		case "redpanda":
			workerClient = redpandaworker.NewRedpandaClient(
				config.App.Worker.Redpanda.Brokers,
				config.App.Worker.Redpanda.Topic,
			)
		default:
			workerClient = infraworker.NewNoOpClient()
		}

		// Initialize worker server
		switch featureFlag.Worker.Backend {
		case "asynq":
			workerServer = asynqworker.NewAsynqServer(
				config.App.Worker.Asynq.RedisURL,
				config.App.Worker.Asynq.Concurrency,
			)
		case "rabbitmq":
			if server, err := rabbitmqworker.NewRabbitMQServer(
				config.App.Worker.RabbitMQ.URL,
				config.App.Worker.RabbitMQ.Exchange,
				config.App.Worker.RabbitMQ.Queue,
				config.App.Worker.RabbitMQ.PrefetchCount,
			); err == nil {
				workerServer = server
			} else {
				workerServer = infraworker.NewNoOpServer()
			}
		case "redpanda":
			workerServer = redpandaworker.NewRedpandaServer(
				config.App.Worker.Redpanda.Brokers,
				config.App.Worker.Redpanda.Topic,
				config.App.Worker.Redpanda.ConsumerGroup,
				config.App.Worker.Redpanda.WorkerCount,
			)
		default:
			workerServer = infraworker.NewNoOpServer()
		}
	} else {
		// Use no-op implementations when workers are disabled
		workerClient = infraworker.NewNoOpClient()
		workerServer = infraworker.NewNoOpServer()
	}

	// Initialize storage service (before modules that depend on it)
	var storageService storage.StorageService
	if featureFlag.Storage.Enabled && featureFlag.Storage.Backend != "noop" && config != nil {
		switch featureFlag.Storage.Backend {
		case "local":
			if svc, err := local.NewLocalStorageService(local.LocalStorageConfig{
				BasePath:          config.App.Storage.Local.BasePath,
				MaxFileSize:       config.App.Storage.Local.MaxFileSize,
				AllowPublicAccess: config.App.Storage.Local.AllowPublicAccess,
				PublicURL:         config.App.Storage.Local.PublicURL,
				CreateMissingDirs: true,
			}); err == nil {
				storageService = svc
			} else {
				storageService = noop.NewNoOpStorageService()
			}
		case "s3", "s3-compatible":
			if svc, err := s3.NewS3StorageService(s3.S3StorageConfig{
				Region:               config.App.Storage.S3.Region,
				Bucket:               config.App.Storage.S3.Bucket,
				AccessKeyID:          config.App.Storage.S3.AccessKeyID,
				SecretAccessKey:      config.App.Storage.S3.SecretAccessKey,
				Endpoint:             config.App.Storage.S3.Endpoint,
				UseSSL:               config.App.Storage.S3.UseSSL,
				PathStyle:            config.App.Storage.S3.PathStyle,
				PresignedURLTTL:      featureFlag.Storage.S3.PresignedURLTTL,
				ServerSideEncryption: featureFlag.Storage.S3.EnableEncryption,
				StorageClass:         featureFlag.Storage.S3.StorageClass,
			}); err == nil {
				storageService = svc
			} else {
				storageService = noop.NewNoOpStorageService()
			}
		case "gcs":
			if svc, err := gcs.NewGCSStorageService(context.Background(), gcs.GCSStorageConfig{
				ProjectID:       config.App.Storage.GCS.ProjectID,
				Bucket:          config.App.Storage.GCS.Bucket,
				CredentialsFile: config.App.Storage.GCS.CredentialsFile,
				CredentialsJSON: config.App.Storage.GCS.CredentialsJSON,
				StorageClass:    featureFlag.Storage.GCS.StorageClass,
				Location:        config.App.Storage.GCS.Location,
				MetadataCache:   featureFlag.Storage.GCS.MetadataCache,
			}); err == nil {
				storageService = svc
			} else {
				storageService = noop.NewNoOpStorageService()
			}
		default:
			storageService = noop.NewNoOpStorageService()
		}
	} else {
		// Use no-op implementation when storage is disabled
		storageService = noop.NewNoOpStorageService()
	}

	return &Container{
		Cache:             cacheInstance,
		EventBus:          eventBus,
		EmailClient:       emailService,
		StorageService:    storageService,
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
		WorkerClient:      workerClient,
		WorkerServer:      workerServer,
	}
}
