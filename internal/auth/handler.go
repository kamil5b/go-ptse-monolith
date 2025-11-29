package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c echo.Context) error {
	var r registerReq
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	u, err := h.svc.Register(r.Email, r.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	// in real app send activation email. We return token for dev convenience
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": u.ID, "activation_token": u.ActivationToken})
}

func (h *Handler) Activate(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, "missing token")
	}
	if err := h.svc.Activate(token); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "activated")
}

type loginReq struct{ Email, Password string }

func (h *Handler) Login(c echo.Context) error {
	var r loginReq
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	u, err := h.svc.Authenticate(r.Email, r.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}
	token, _ := h.svc.GenerateToken(u.ID, 24*time.Hour)
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

type forgotReq struct{ Email string }

func (h *Handler) ForgotPassword(c echo.Context) error {
	var r forgotReq
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	token, err := h.svc.ForgotPassword(r.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	// In prod send email; return token for dev
	return c.JSON(http.StatusOK, map[string]string{"reset_token": token})
}

type resetReq struct{ Token, NewPassword string }

func (h *Handler) ResetPassword(c echo.Context) error {
	var r resetReq
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.ResetPassword(r.Token, r.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "ok")
}
