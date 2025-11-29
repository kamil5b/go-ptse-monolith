package product

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type ProductService interface {
	Create(*Product) error
	Get(string) (*Product, error)
	List() ([]Product, error)
	Update(*Product) error
	Delete(id, by string) error
}

type Handler struct {
	svc ProductService
}

func NewHandler(s ProductService) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Create(c echo.Context) error {
	var req Product
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if uid := c.Get("user_id"); uid != nil {
		if s, ok := uid.(string); ok {
			req.CreatedBy = s
		}
	}
	if err := h.svc.Create(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, req)
}

func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")
	p, err := h.svc.Get(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) List(c echo.Context) error {
	lst, err := h.svc.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, lst)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	var req Product
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	req.ID = id
	if uid := c.Get("user_id"); uid != nil {
		if s, ok := uid.(string); ok {
			req.UpdatedBy = &s
		}
	}
	now := time.Now().UTC()
	req.UpdatedAt = &now
	if err := h.svc.Update(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, req)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	by := ""
	if uid := c.Get("user_id"); uid != nil {
		if s, ok := uid.(string); ok {
			by = s
		}
	}
	if err := h.svc.Delete(id, by); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
