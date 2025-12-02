package fasthttp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

type FastHTTPContext struct {
	ctx *fasthttp.RequestCtx
}

func (c FastHTTPContext) BindJSON(obj any) error {
	if len(c.ctx.PostBody()) == 0 {
		return errors.New("empty body")
	}
	return json.Unmarshal(c.ctx.PostBody(), obj)
}
func (c FastHTTPContext) BindURI(obj any) error    { return nil }
func (c FastHTTPContext) BindQuery(obj any) error  { return nil }
func (c FastHTTPContext) BindHeader(obj any) error { return nil }
func (c FastHTTPContext) Bind(obj any) error       { return nil }
func (c FastHTTPContext) JSON(code int, v any) error {
	c.ctx.SetContentType("application/json")
	c.ctx.SetStatusCode(code)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = c.ctx.Write(b)
	return err
}
func (c FastHTTPContext) Param(n string) string {
	v := c.ctx.UserValue(n)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func (c FastHTTPContext) GetUserID() string           { return "" }
func (c FastHTTPContext) Get(key string) any          { return nil }
func (c FastHTTPContext) Set(key string, value any)   {}
func (c FastHTTPContext) GetContext() context.Context { return context.Background() }
func (c FastHTTPContext) GetHeader(key string) string { return string(c.ctx.Request.Header.Peek(key)) }
func (c FastHTTPContext) SetHeader(key, value string) { c.ctx.Response.Header.Set(key, value) }
func (c FastHTTPContext) GetCookie(name string) (string, error) {
	v := c.ctx.Request.Header.Cookie(name)
	if len(v) == 0 {
		return "", nil
	}
	return string(v), nil
}
func (c FastHTTPContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := fasthttp.Cookie{}
	cookie.SetKey(name)
	cookie.SetValue(value)
	cookie.SetPath(path)
	cookie.SetDomain(domain)
	cookie.SetSecure(secure)
	cookie.SetHTTPOnly(httpOnly)
	cookie.SetExpire(time.Now().Add(time.Duration(maxAge) * time.Second))
	c.ctx.Response.Header.SetCookie(&cookie)
}
func (c FastHTTPContext) RemoveCookie(name string) {
	cookie := fasthttp.Cookie{}
	cookie.SetKey(name)
	cookie.SetValue("")
	cookie.SetPath("/")
	cookie.SetExpire(time.Now().Add(-1 * time.Hour))
	c.ctx.Response.Header.SetCookie(&cookie)
}
func (c FastHTTPContext) GetClientIP() string  { return c.ctx.RemoteAddr().String() }
func (c FastHTTPContext) GetUserAgent() string { return string(c.ctx.Request.Header.UserAgent()) }

func NewFastHTTPContext(ctx *fasthttp.RequestCtx) FastHTTPContext { return FastHTTPContext{ctx: ctx} }
