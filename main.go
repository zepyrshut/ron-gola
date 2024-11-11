package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Context is the context of the current HTTP request.
type Context struct {
	Ctx context.Context
	W   http.ResponseWriter
	R   *http.Request
}

// JSON serializes the given data to JSON and writes it to the response.
func (c *Context) JSON(statusCode int, data interface{}) {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(statusCode)
	if err := json.NewEncoder(c.W).Encode(data); err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
	}
}

// HandlerFunc defines the handler used by the framework.
type HandlerFunc func(c *Context)

// Engine is the framework's instance, it contains the router and middleware.
type Engine struct {
	router *RouterGroup
}

// RouterGroup is used internally to configure router, associated with a prefix and an array of handlers (middlewares).
type RouterGroup struct {
	Handlers []HandlerFunc
	engine   *Engine
}

// New creates a new Engine instance.
func New() *Engine {
	engine := &Engine{}
	engine.router = &RouterGroup{engine: engine}
	return engine
}

// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	engine.router.handleHTTPRequest(ctx, w, r)
}

// handleHTTPRequest handles the HTTP request.
func (group *RouterGroup) handleHTTPRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	c := &Context{Ctx: ctx, W: w, R: r}
	for _, handler := range group.Handlers {
		handler(c)
	}
}

// Use adds middleware to the router group.
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.Handlers = append(group.Handlers, middleware...)
}

// GET adds a GET route to the router group.
func (group *RouterGroup) GET(path string, handler HandlerFunc) {
	// Implement route registration logic here
	group.Handlers = append(group.Handlers, handler)
}

func main() {
	engine := New()

	engine.router.GET("/", sayHelloHandler)
	engine.router.GET("/json", sayHelloJSONHandler)

	slog.Info("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", engine)
}

func sayHelloHandler(c *Context) {
	slog.Info("called sayHelloHandler")
	c.W.Write([]byte("Hello, World!"))
}

type Something struct {
	Name string `json:"name"`
}

func sayHelloJSONHandler(c *Context) {
	slog.Info("called sayHelloJSONHandler")
	something := Something{Name: "something"}
	c.JSON(http.StatusOK, something)
}
