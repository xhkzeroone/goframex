package ginx

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GetHandler interface {
	Handle(ctx *Context, headers, query, path map[string]string) error
}

type PostHandler interface {
	Handle(ctx *Context, body []byte, headers, query, path map[string]string) error
}

type Context struct {
	*gin.Context
	bodyBytes []byte
	path      string
	method    string
	query     map[string]string
	headers   map[string]string
	pathVar   map[string]string
	status    int
	response  any
}

func (c *Context) Body() []byte               { return c.bodyBytes }
func (c *Context) Method() string             { return c.method }
func (c *Context) Path() string               { return c.path }
func (c *Context) Query() map[string]string   { return c.query }
func (c *Context) Headers() map[string]string { return c.headers }
func (c *Context) PathVar() map[string]string { return c.pathVar }
func (c *Context) Status() int                { return c.status }
func (c *Context) Response() any              { return c.response }

func (c *Context) JSON(status int, data any) {
	c.status = status
	c.response = data
}

func (c *Context) Bind(data any) error {
	return json.Unmarshal(c.bodyBytes, data)
}

type HandlerFunc func(*Context) error
type Middleware func(HandlerFunc) HandlerFunc

type MiddlewareChain struct {
	middlewares []Middleware
}

func (m *MiddlewareChain) Use(mw ...Middleware) {
	m.middlewares = append(m.middlewares, mw...)
}

func (m *MiddlewareChain) Wrap(handler HandlerFunc) gin.HandlerFunc {
	final := handler
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		final = m.middlewares[i](final)
	}

	return func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()

		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		query := make(map[string]string)
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				query[k] = v[0]
			}
		}

		pathVar := make(map[string]string)
		for _, p := range c.Params {
			pathVar[p.Key] = p.Value
		}

		myCtx := &Context{
			Context:   c,
			method:    c.Request.Method,
			path:      c.Request.URL.Path,
			bodyBytes: body,
			headers:   headers,
			query:     query,
			pathVar:   pathVar,
			status:    http.StatusInternalServerError,
		}

		// Gọi handler
		if err := final(myCtx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if myCtx.Response() != nil {
			c.JSON(myCtx.Status(), myCtx.Response())
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

func Compose(chains ...*MiddlewareChain) *MiddlewareChain {
	out := &MiddlewareChain{}
	for _, c := range chains {
		out.middlewares = append(out.middlewares, c.middlewares...)
	}
	return out
}

type Route struct {
	Path       string
	Method     string
	Handler    HandlerFunc
	Middleware []Middleware
}

type RouterGroup struct {
	group      *gin.RouterGroup
	Middleware *MiddlewareChain
	Parent     *RouterGroup
}

func (g *RouterGroup) Use(mws ...Middleware) {
	g.Middleware.Use(mws...)
}

func (g *RouterGroup) Group(path string, middleware ...Middleware) *RouterGroup {
	sub := g.group.Group(path)
	return &RouterGroup{
		group: sub,
		Middleware: &MiddlewareChain{
			middlewares: middleware,
		},
		Parent: g,
	}
}

func (g *RouterGroup) GET(path string, handler GetHandler, middleware ...Middleware) {
	chain := &MiddlewareChain{}
	chain.Use(middleware...)
	final := Compose(g.collectMiddleware(), chain)
	g.group.GET(path, final.Wrap(AdaptGetHandler(handler)))
}

func (g *RouterGroup) POST(path string, handler PostHandler, middleware ...Middleware) {
	chain := &MiddlewareChain{}
	chain.Use(middleware...)
	final := Compose(g.collectMiddleware(), chain)
	g.group.POST(path, final.Wrap(AdaptPostHandler(handler)))
}

func (g *RouterGroup) PUT(path string, handler PostHandler, middleware ...Middleware) {
	chain := &MiddlewareChain{}
	chain.Use(middleware...)
	final := Compose(g.collectMiddleware(), chain)
	g.group.PUT(path, final.Wrap(AdaptPostHandler(handler)))
}

func (g *RouterGroup) DELETE(path string, handler PostHandler, middleware ...Middleware) {
	chain := &MiddlewareChain{}
	chain.Use(middleware...)
	final := Compose(g.collectMiddleware(), chain)
	g.group.DELETE(path, final.Wrap(AdaptPostHandler(handler)))
}

func (g *RouterGroup) collectMiddleware() *MiddlewareChain {
	var all []*MiddlewareChain
	curr := g
	for curr != nil {
		all = append([]*MiddlewareChain{curr.Middleware}, all...)
		curr = curr.Parent
	}
	return Compose(all...)
}

func registerRoutes(group *RouterGroup, routes []Route) {
	for _, r := range routes {
		chain := &MiddlewareChain{}
		chain.Use(r.Middleware...)
		final := Compose(group.collectMiddleware(), chain)
		switch r.Method {
		case http.MethodGet:
			group.group.GET(r.Path, final.Wrap(r.Handler))
		case http.MethodPost:
			group.group.POST(r.Path, final.Wrap(r.Handler))
		case http.MethodPut:
			group.group.PUT(r.Path, final.Wrap(r.Handler))
		case http.MethodDelete:
			group.group.DELETE(r.Path, final.Wrap(r.Handler))
		default:
			panic("unsupported method: " + r.Method)
		}
	}
}

type Server struct {
	*gin.Engine
	rootGroup  *RouterGroup
	config     *Config
	httpServer *http.Server
}

func New(cfg *Config) *Server {
	gin.SetMode(cfg.Mode)
	engine := gin.Default()
	rootGroup := &RouterGroup{
		Middleware: &MiddlewareChain{},
		group:      engine.Group(cfg.RootPath),
	}

	return &Server{
		Engine:    engine,
		rootGroup: rootGroup,
		config:    cfg,
	}
}

func (s *Server) Start() error {
	addr := s.config.GetAddr()
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}
	log.Printf("Server is running at %s", addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down ginx...")
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Use(middleware ...Middleware) {
	s.rootGroup.Use(middleware...)
}

func (s *Server) Group(relativePath string, middleware ...Middleware) *RouterGroup {
	return s.rootGroup.Group(relativePath, middleware...)
}

func (s *Server) Routes(routes ...Route) {
	registerRoutes(s.rootGroup, routes)
}

func (s *Server) RoutesGroup(relativePath string, routes ...Route) {
	group := s.rootGroup.Group(relativePath)
	registerRoutes(group, routes)
}

func (s *Server) GET(relativePath string, handler GetHandler, middleware ...Middleware) {
	s.rootGroup.GET(relativePath, handler, middleware...)
}

func (s *Server) POST(relativePath string, handler PostHandler, middleware ...Middleware) {
	s.rootGroup.POST(relativePath, handler, middleware...)
}

func (s *Server) PUT(relativePath string, handler PostHandler, middleware ...Middleware) {
	s.rootGroup.PUT(relativePath, handler, middleware...)
}

func (s *Server) DELETE(relativePath string, handler PostHandler, middleware ...Middleware) {
	s.rootGroup.DELETE(relativePath, handler, middleware...)
}

type HealthCheckFunc interface {
	Liveness() (bool, error)
	Readiness() (bool, error)
	Terminate() (bool, error)
}

type DefaultHealthCheckFunc struct {
}

func (r *DefaultHealthCheckFunc) Liveness() (bool, error) {
	return true, nil
}

func (r *DefaultHealthCheckFunc) Readiness() (bool, error) {
	return true, nil
}

func (r *DefaultHealthCheckFunc) Terminate() (bool, error) {
	return true, nil
}

func (s *Server) HealthCheck() {
	s.HealthCheckWithFunc(&DefaultHealthCheckFunc{})
}

func (s *Server) HealthCheckWithFunc(healthCheck HealthCheckFunc) {
	s.Engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	s.Engine.GET("/liveness", func(c *gin.Context) {
		ok, err := healthCheck.Liveness()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	s.Engine.GET("/readiness", func(c *gin.Context) {
		ok, err := healthCheck.Readiness()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	s.Engine.POST("/terminate", func(c *gin.Context) {
		ok, err := healthCheck.Terminate()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "cannot terminate"})
			return
		}

		// Gửi response trước, rồi tắt server sau 1 giây
		c.JSON(http.StatusOK, gin.H{"status": "terminating"})
		go func() {
			time.Sleep(1 * time.Second)
			_ = s.Stop(context.Background())
		}()
	})
}

func AdaptPostHandler(h PostHandler) HandlerFunc {
	return func(c *Context) error {
		err := h.Handle(c, c.Body(), c.Headers(), c.Query(), c.PathVar())
		return err
	}
}

func AdaptGetHandler(h GetHandler) HandlerFunc {
	return func(c *Context) error {
		err := h.Handle(c, c.Headers(), c.Query(), c.PathVar())
		return err
	}
}
