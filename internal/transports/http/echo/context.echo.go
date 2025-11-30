package echo

import (
	"context"

	"github.com/labstack/echo/v4"
)

type EchoContext struct {
	c echo.Context
}

func (ctx EchoContext) BindJSON(obj any) error   { return ctx.c.Bind(obj) }
func (ctx EchoContext) BindURI(obj any) error    { return ctx.c.Bind(obj) }
func (ctx EchoContext) BindQuery(obj any) error  { return ctx.c.Bind(obj) }
func (ctx EchoContext) BindHeader(obj any) error { return ctx.c.Bind(obj) }
func (ctx EchoContext) Bind(obj any) error       { return ctx.c.Bind(obj) }
func (ctx EchoContext) JSON(code int, v any) error {
	return ctx.c.JSON(code, v)
}
func (ctx EchoContext) Param(n string) string {
	return ctx.c.Param(n)
}
func (ctx EchoContext) GetUserID() string {
	return ctx.c.Get("user_id").(string)
}
func (ctx EchoContext) Get(key string) any {
	return ctx.c.Get(key)
}
func (ctx EchoContext) GetContext() context.Context {
	return ctx.c.Request().Context()
}

func NewEchoContext(c echo.Context) EchoContext {
	return EchoContext{c: c}
}
