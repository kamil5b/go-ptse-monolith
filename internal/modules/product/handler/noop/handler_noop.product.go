package noop

import (
	"net/http"

	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

type Handler struct{}

func NewUnimplementedHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Create(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Get(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) List(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Update(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}

func (h *Handler) Delete(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "unimplemented"})
}
