package nethttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type NetHTTPContext struct {
	w http.ResponseWriter
	r *http.Request
}

func (ctx NetHTTPContext) BindJSON(obj any) error {
	if ctx.r.Body == nil {
		return errors.New("empty body")
	}
	defer ctx.r.Body.Close()
	return json.NewDecoder(ctx.r.Body).Decode(obj)
}
func (ctx NetHTTPContext) BindURI(obj any) error {
	// not implementing direct binding to struct; leave as noop
	return nil
}
func (ctx NetHTTPContext) BindQuery(obj any) error {
	return nil
}
func (ctx NetHTTPContext) BindHeader(obj any) error { return nil }
func (ctx NetHTTPContext) Bind(obj any) error       { return nil }
func (ctx NetHTTPContext) JSON(code int, v any) error {
	ctx.w.Header().Set("Content-Type", "application/json")
	ctx.w.WriteHeader(code)
	return json.NewEncoder(ctx.w).Encode(v)
}
func (ctx NetHTTPContext) Param(n string) string {
	vars := mux.Vars(ctx.r)
	return vars[n]
}
func (ctx NetHTTPContext) GetUserID() string           { return "" }
func (ctx NetHTTPContext) Get(key string) any          { return nil }
func (ctx NetHTTPContext) Set(key string, value any)   {}
func (ctx NetHTTPContext) GetContext() context.Context { return ctx.r.Context() }
func (ctx NetHTTPContext) GetHeader(key string) string { return ctx.r.Header.Get(key) }
func (ctx NetHTTPContext) SetHeader(key, value string) { ctx.w.Header().Set(key, value) }
func (ctx NetHTTPContext) GetCookie(name string) (string, error) {
	c, err := ctx.r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}
		return "", err
	}
	return c.Value, nil
}
func (ctx NetHTTPContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	http.SetCookie(ctx.w, cookie)
}
func (ctx NetHTTPContext) RemoveCookie(name string) {
	cookie := &http.Cookie{Name: name, Value: "", MaxAge: -1, Path: "/"}
	http.SetCookie(ctx.w, cookie)
}
func (ctx NetHTTPContext) GetClientIP() string  { return ctx.r.RemoteAddr }
func (ctx NetHTTPContext) GetUserAgent() string { return ctx.r.UserAgent() }

func NewNetHTTPContext(w http.ResponseWriter, r *http.Request) NetHTTPContext {
	return NetHTTPContext{w: w, r: r}
}
