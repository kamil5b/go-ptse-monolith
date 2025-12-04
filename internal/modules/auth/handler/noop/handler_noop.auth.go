package noop

import (
	"net/http"

	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

type NoopHandler struct{}

func NewNoopHandler() *NoopHandler {
	return &NoopHandler{}
}

func (h *NoopHandler) Login(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) Register(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) Logout(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) RefreshToken(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) ValidateToken(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) ChangePassword(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) GetProfile(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) GetSessions(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) RevokeSession(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}

func (h *NoopHandler) RevokeAllSessions(c sharedctx.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "auth not implemented"})
}
