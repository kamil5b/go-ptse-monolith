package core

import (
	"go-modular-monolith/internal/domain/product"

	repoMongo "go-modular-monolith/internal/modules/product/repository/mongo"
	repoSQL "go-modular-monolith/internal/modules/product/repository/sql"

	serviceUnimplemented "go-modular-monolith/internal/modules/product/service/noop"
	serviceV1 "go-modular-monolith/internal/modules/product/service/v1"

	handlerUnimplemented "go-modular-monolith/internal/modules/product/handler/noop"
	handlerV1 "go-modular-monolith/internal/modules/product/handler/v1"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	ProductRepository product.ProductRepository
	ProductService    product.ProductService
	ProductHandler    product.ProductHandler
}

func NewContainer(
	featureFlag FeatureFlag,
	db *sqlx.DB,
	mongo *mongo.Client,
) *Container {
	var (
		productRepository product.ProductRepository
		productService    product.ProductService
		productHandler    product.ProductHandler
	)

	// repo
	switch featureFlag.Repository.Product {
	case "mongo":
		productRepository = repoMongo.NewMongoRepository(mongo, "appdb")
	case "postgres":
		productRepository = repoSQL.NewSQLRepository(db)
	default:
		// productRepository = repoUnimplemented.NewUnimplementedRepository()
	}

	// service
	switch featureFlag.Service.Product {
	case "v1":
		productService = serviceV1.NewServiceV1(productRepository)
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

	return &Container{
		ProductService: productService,
		ProductHandler: productHandler,
	}
}
