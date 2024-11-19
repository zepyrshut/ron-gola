package ron

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type (
	EngineOptions func(*Engine)

	Data map[string]any

	Context struct {
		C context.Context
		W http.ResponseWriter
		R *http.Request
		E *Engine
	}

	router struct {
		path    string
		method  string
		handler func(*Context)
		group   *routerGroup
	}

	routerGroup struct {
		prefix      string
		middlewares middlewareChain
		engine      *Engine
	}

	middlewareChain []func(http.Handler) http.Handler

	Engine struct {
		LogLevel    slog.Level
		Render      *Render
		mux         *http.ServeMux
		middlewares middlewareChain
		router      []router
	}
)

func defaultEngine() *Engine {
	return &Engine{
		mux:      http.NewServeMux(),
		LogLevel: slog.LevelInfo,
	}
}

func New(opts ...EngineOptions) *Engine {
	config := defaultEngine()
	return config.apply(opts...)
}

func (e *Engine) apply(opts ...EngineOptions) *Engine {
	for _, opt := range opts {
		if opt != nil {
			opt(e)
		}
	}

	return e
}

func (e *Engine) applyMiddlewares(handler http.Handler) http.Handler {
	for i := len(e.middlewares) - 1; i >= 0; i-- {
		handler = e.middlewares[i](handler)
	}
	return handler
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := e.applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range e.router {
			if strings.HasPrefix(r.URL.Path, route.path) && r.Method == route.method {
				groupHandler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					route.handler(&Context{W: w, R: r, E: e})
				}))
				if route.group != nil {
					for i := len(route.group.middlewares) - 1; i >= 0; i-- {
						groupHandler = route.group.middlewares[i](groupHandler)
					}
				}
				groupHandler.ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}))
	handler.ServeHTTP(w, r)
}

func (e *Engine) Run(addr string) error {
	newLogger(e.LogLevel)
	return http.ListenAndServe(addr, e)
}

func (e *Engine) handleRequest(w http.ResponseWriter, r *http.Request) {
	e.mux.ServeHTTP(w, r)
}

func (e *Engine) USE(middleware ...func(http.Handler) http.Handler) {
	e.middlewares = append(e.middlewares, middleware...)
}

func (e *Engine) GET(path string, handler func(*Context)) {
	e.router = append(e.router, router{path: path, method: http.MethodGet, handler: handler})
}

func (e *Engine) POST(path string, handler func(*Context)) {
	e.router = append(e.router, router{path: path, method: http.MethodPost, handler: handler})
}

func (e *Engine) GROUP(prefix string) *routerGroup {
	return &routerGroup{
		prefix: prefix,
		engine: e,
	}
}

func (rg *routerGroup) USE(middleware ...func(http.Handler) http.Handler) {
	rg.middlewares = append(rg.middlewares, middleware...)
}

func (rg *routerGroup) GET(path string, handler func(*Context)) {
	rg.engine.router = append(rg.engine.router, router{path: rg.prefix + path, method: http.MethodGet, handler: handler, group: rg})
}

func (rg *routerGroup) POST(path string, handler func(*Context)) {
	rg.engine.router = append(rg.engine.router, router{path: rg.prefix + path, method: http.MethodPost, handler: handler, group: rg})
}

// Static serves static files from a specified directory, accessible through a defined URL path.
//
// The `path` parameter represents the URL prefix to access the static files.
// The `dir` parameter represents the actual filesystem path where the static files are located.
//
// Example:
// Calling r.Static("assets", "./folder") will make the contents of the "./folder" directory
// accessible in the browser at "/assets". For instance, a file located at "./folder/image.png"
// would be available at "/assets/image.png" in HTML templates.
func (e *Engine) Static(path, dir string) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	if !strings.HasPrefix(dir, "./") {
		dir = "./" + dir
	}

	fs := http.FileServer(http.Dir(dir))
	e.GET(path, func(c *Context) {
		http.StripPrefix(path, fs).ServeHTTP(c.W, c.R)
	})
	slog.Info("Static files served", "path", path, "dir", dir)
}

func (c *Context) JSON(code int, data any) {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(c.W)
	if err := encoder.Encode(data); err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) HTML(code int, name string, td *TemplateData) {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.E.Render.Template(c.W, name, td)
	if err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
	}
}

func newLogger(level slog.Level) {
	now := time.Now().Format("2006-01-02")
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}
	f, _ := os.OpenFile(fmt.Sprintf("logs/log%s.log", now), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	mw := io.MultiWriter(os.Stdout, f)

	logger := slog.New(slog.NewTextHandler(mw, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))

	slog.SetDefault(logger)
}
