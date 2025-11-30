package v1

import (
	"go-modular-monolith/internal/domain/product"
	"net/http"
)

type Handler struct {
	svc product.ProductService
}

func NewHandler(s product.ProductService) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Create(c product.Context) error {
	var req product.CreateProductRequest
	ctx := c.GetContext()
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	createdBy := c.GetUserID()
	p, err := h.svc.Create(ctx, &req, createdBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) Get(c product.Context) error {
	ctx := c.GetContext()
	id := c.Param("id")
	p, err := h.svc.Get(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) List(c product.Context) error {
	ctx := c.GetContext()
	lst, err := h.svc.List(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, lst)
}

func (h *Handler) Update(c product.Context) error {
	ctx := c.GetContext()
	id := c.Param("id")
	var req product.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	req.ID = id
	updatedBy := c.GetUserID()
	p, err := h.svc.Update(ctx, &req, updatedBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) Delete(c product.Context) error {
	ctx := c.GetContext()
	id := c.Param("id")
	by := ""
	if uid := c.Get("user_id"); uid != nil {
		if s, ok := uid.(string); ok {
			by = s
		}
	}
	if err := h.svc.Delete(ctx, id, by); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
