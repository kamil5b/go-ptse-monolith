package http

import (
	"go-modular-monolith/internal/domain/product"
	"go-modular-monolith/pkg/routes"
)

func NewRoutes(
	productHandler product.ProductHandler,
) *[]routes.Route {
	return &[]routes.Route{
		{
			Method:  "GET",
			Path:    "/product",
			Handler: productHandler.List,
		},
		{
			Method:  "POST",
			Path:    "/product",
			Handler: productHandler.Create,
		},
	}
}
