package noop

import (
	"go-modular-monolith/internal/domain/product"
	"net/http"
)

type Handler struct{}

func NewUnimplementedHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Create(c product.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Get(c product.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) List(c product.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Update(c product.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Delete(c product.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}
