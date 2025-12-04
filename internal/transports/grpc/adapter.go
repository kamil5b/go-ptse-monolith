package grpctransport

import (
	"context"
	"errors"

	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

// GRPCContext is a lightweight adapter that implements the sharedctx.Context
// interface for gRPC handlers. It provides minimal behavior expected by
// handlers that are designed to be framework-agnostic. For fields or
// operations not applicable to gRPC (cookies, path params) it returns
// appropriate zero values or errors.
type GRPCContext struct {
	ctx     context.Context
	headers map[string]string
	vals    map[string]interface{}
}

// NewGRPCContext creates a new GRPCContext wrapping a context.Context.
func NewGRPCContext(ctx context.Context, headers map[string]string) *GRPCContext {
	if headers == nil {
		headers = map[string]string{}
	}
	return &GRPCContext{ctx: ctx, headers: headers, vals: map[string]interface{}{}}
}

// Bind methods are not meaningful for gRPC; handlers should decode from
// protobuf messages. Return an error to indicate unsupported operation.
func (g *GRPCContext) BindJSON(obj any) error  { return errors.New("BindJSON not supported for gRPC") }
func (g *GRPCContext) BindURI(obj any) error   { return errors.New("BindURI not supported for gRPC") }
func (g *GRPCContext) BindQuery(obj any) error { return errors.New("BindQuery not supported for gRPC") }
func (g *GRPCContext) BindHeader(obj any) error {
	return errors.New("BindHeader not supported for gRPC")
}
func (g *GRPCContext) Bind(obj any) error { return errors.New("Bind not supported for gRPC") }

// JSON is not used in gRPC; gRPC handlers should return protobuf responses.
func (g *GRPCContext) JSON(code int, v any) error {
	return errors.New("JSON response not supported for gRPC; return protobuf message instead")
}

// Param (path param) is not available for gRPC
func (g *GRPCContext) Param(name string) string { return "" }

// GetUserID attempts to read a user id set in values, else empty string.
func (g *GRPCContext) GetUserID() string {
	if v, ok := g.vals["user_id"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (g *GRPCContext) Get(key string) any          { return g.vals[key] }
func (g *GRPCContext) Set(key string, value any)   { g.vals[key] = value }
func (g *GRPCContext) GetContext() context.Context { return g.ctx }

func (g *GRPCContext) GetHeader(key string) string        { return g.headers[key] }
func (g *GRPCContext) SetHeader(key string, value string) { g.headers[key] = value }

// Cookies are not applicable in gRPC
func (g *GRPCContext) GetCookie(name string) (string, error) {
	return "", errors.New("cookies not supported for gRPC")
}
func (g *GRPCContext) SetCookie(name string, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	// no-op
}
func (g *GRPCContext) RemoveCookie(name string) {
	// no-op
}

func (g *GRPCContext) GetClientIP() string  { return g.headers["x-forwarded-for"] }
func (g *GRPCContext) GetUserAgent() string { return g.headers["user-agent"] }

// Ensure GRPCContext implements sharedctx.Context
var _ sharedctx.Context = (*GRPCContext)(nil)
