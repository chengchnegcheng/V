package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router 是HTTP路由器接口
type Router interface {
	http.Handler
	Group(path string) RouterGroup
}

// RouterGroup 是路由组接口
type RouterGroup interface {
	GET(path string, handlers ...gin.HandlerFunc)
	POST(path string, handlers ...gin.HandlerFunc)
	PUT(path string, handlers ...gin.HandlerFunc)
	DELETE(path string, handlers ...gin.HandlerFunc)
}

// Context 是HTTP上下文接口
type Context = gin.Context

// ginRouter 是gin实现的路由器
type ginRouter struct {
	engine *gin.Engine
}

// ginRouterGroup 是gin实现的路由组
type ginRouterGroup struct {
	group *gin.RouterGroup
}

// NewRouter 创建一个新的路由器
func NewRouter() Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	return &ginRouter{engine: engine}
}

// ServeHTTP 实现http.Handler接口
func (r *ginRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// Group 创建一个新的路由组
func (r *ginRouter) Group(path string) RouterGroup {
	return &ginRouterGroup{group: r.engine.Group(path)}
}

// GET 注册GET路由
func (g *ginRouterGroup) GET(path string, handlers ...gin.HandlerFunc) {
	g.group.GET(path, handlers...)
}

// POST 注册POST路由
func (g *ginRouterGroup) POST(path string, handlers ...gin.HandlerFunc) {
	g.group.POST(path, handlers...)
}

// PUT 注册PUT路由
func (g *ginRouterGroup) PUT(path string, handlers ...gin.HandlerFunc) {
	g.group.PUT(path, handlers...)
}

// DELETE 注册DELETE路由
func (g *ginRouterGroup) DELETE(path string, handlers ...gin.HandlerFunc) {
	g.group.DELETE(path, handlers...)
}
