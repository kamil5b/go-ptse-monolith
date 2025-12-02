package v1

import (
	"go-modular-monolith/internal/modules/user/domain"
	sharedctx "go-modular-monolith/internal/shared/context"
	"net/http"
)

type Handler struct {
	svc domain.Service
}

func NewHandler(s domain.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Create(c sharedctx.Context) error {
	var req domain.CreateUserRequest
	ctx := c.GetContext()
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	createdBy := c.GetUserID()
	u, err := h.svc.Create(ctx, &req, createdBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Get(c sharedctx.Context) error {
	ctx := c.GetContext()
	id := c.Param("id")
	u, err := h.svc.Get(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) List(c sharedctx.Context) error {
	ctx := c.GetContext()
	lst, err := h.svc.List(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, lst)
}

func (h *Handler) Update(c sharedctx.Context) error {
	ctx := c.GetContext()
	id := c.Param("id")
	var req domain.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	req.ID = id
	updatedBy := c.GetUserID()
	u, err := h.svc.Update(ctx, &req, updatedBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) Delete(c sharedctx.Context) error {
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
