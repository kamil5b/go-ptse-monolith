package gin

import (
	"context"

	"github.com/gin-gonic/gin"
)

type GinContext struct {
	c *gin.Context
}

func (ctx GinContext) BindJSON(obj any) error   { return ctx.c.ShouldBindJSON(obj) }
func (ctx GinContext) BindURI(obj any) error    { return ctx.c.ShouldBindUri(obj) }
func (ctx GinContext) BindQuery(obj any) error  { return ctx.c.ShouldBindQuery(obj) }
func (ctx GinContext) BindHeader(obj any) error { return ctx.c.ShouldBindHeader(obj) }
func (ctx GinContext) Bind(obj any) error       { return ctx.c.ShouldBind(obj) }
func (ctx GinContext) JSON(code int, v any) error {
	ctx.c.JSON(code, v)
	return nil
}
func (ctx GinContext) Param(n string) string {
	return ctx.c.Param(n)
}
func (ctx GinContext) GetUserID() string {
	val, _ := ctx.c.Get("user_id")
	return val.(string)
}
func (ctx GinContext) Get(key string) any {
	val, _ := ctx.c.Get(key)
	return val
}
func (ctx GinContext) GetContext() context.Context {
	return ctx.c.Request.Context()
}

func NewGinContext(c *gin.Context) GinContext {
	return GinContext{c: c}
}
