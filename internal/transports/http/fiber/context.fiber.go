package fiber

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type FiberContext struct {
	c *fiber.Ctx
}

func (f FiberContext) BindJSON(obj any) error {
	return f.c.BodyParser(obj)
}
func (f FiberContext) BindURI(obj any) error    { return nil }
func (f FiberContext) BindQuery(obj any) error  { return nil }
func (f FiberContext) BindHeader(obj any) error { return nil }
func (f FiberContext) Bind(obj any) error       { return nil }
func (f FiberContext) JSON(code int, v any) error {
	f.c.Set("Content-Type", "application/json")
	f.c.Status(code)
	return f.c.JSON(v)
}
func (f FiberContext) Param(n string) string                 { return f.c.Params(n) }
func (f FiberContext) GetUserID() string                     { return "" }
func (f FiberContext) Get(key string) any                    { return f.c.Locals(key) }
func (f FiberContext) Set(key string, value any)             { f.c.Locals(key, value) }
func (f FiberContext) GetContext() context.Context           { return f.c.Context() }
func (f FiberContext) GetHeader(key string) string           { return f.c.Get(key) }
func (f FiberContext) SetHeader(key, value string)           { f.c.Set(key, value) }
func (f FiberContext) GetCookie(name string) (string, error) { return f.c.Cookies(name), nil }
func (f FiberContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := new(fiber.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Path = path
	cookie.Domain = domain
	// leave Expires zero-value to use default behavior
	cookie.HTTPOnly = httpOnly
	cookie.Secure = secure
	f.c.Cookie(cookie)
}
func (f FiberContext) RemoveCookie(name string) { f.c.ClearCookie(name) }
func (f FiberContext) GetClientIP() string      { return f.c.IP() }
func (f FiberContext) GetUserAgent() string     { return f.c.Get("User-Agent") }

func NewFiberContext(c *fiber.Ctx) FiberContext { return FiberContext{c: c} }
