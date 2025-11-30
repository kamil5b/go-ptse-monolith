package http

import (
	"go-modular-monolith/internal/app/core"
)

func NewServer(c *core.Container, featureFlag *core.FeatureFlag) any {
	switch featureFlag.HTTPHandler {
	case "echo":
		return NewEchoServer(c)
	default:
		return NewEchoServer(c)
	}
}
