package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)

// Engine 这个结构体用于实现http.Handler接口
// 整个框架的所有资源都是由Engine统一协调的
type Engine struct {
	//采用组合的方式，使得Engine具有分组的所有方法，也就是说Engine可以作为最顶层的分组
	*RouterGroup
	router *router
	//存储所有的分组
	groups []*RouterGroup
	// html渲染
	htmlTemplates *template.Template
	funcMap       template.FuncMap
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
	// 当我们接收到一个具体请求时，要判断该请求适用于哪些中间件，在这里我们简单通过URL的前缀来判断。
	// 得到中间件列表后，赋值给 c.handlers
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

// Use 将中间件加入group中
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// 检查文件是否存在，并且我们是否有权限
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 将磁盘上的某个文件夹root映射到路由relativePath
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册
	group.GET(urlPattern, handler)
}

// SetFuncMap 自定义渲染函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 加载模板
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// Default 默认使用Logger和Recovery中间件
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
