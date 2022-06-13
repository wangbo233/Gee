package gee

import (
	"log"
	"net/http"
)

type HandlerFunc func(*Context)

// Engine 这个结构体用于实现http.Handler接口
// 整个框架的所有资源都是由Engine统一协调的
type Engine struct {
	*RouterGroup //采用组合的方式，使得Engine具有分组的所有方法，也就是说Engine可以作为最顶层的分组
	router       *router
	groups       []*RouterGroup //存储所有的分组
}

type RouterGroup struct {
	prefix      string        // 路由前缀
	middlewares []HandlerFunc // 支持的中间件
	parent      *RouterGroup  // 当前分组的父亲（为了支持嵌套分组）
	engine      *Engine       // 所有的分组共享一个Engine实例
}

// New 初始化函数，返回一个Engine指针
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group 创建一个新的分组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		engine: engine,
		parent: group,
		prefix: group.prefix + prefix,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// 添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run 启动http服务
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 使用Engine实现http.Handler接口
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
