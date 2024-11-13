package ron

import (
	"context"
	"encoding/json"
	"net/http"
)

type Data map[string]any

type Context struct {
	C context.Context
	W http.ResponseWriter
	R *http.Request
	E *Engine
}

type Engine struct {
	mux      *http.ServeMux
	Renderer *Render
}

func (c *Context) JSON(code int, data any) {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(c.W)
	if err := encoder.Encode(data); err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) HTML(code int, name string, data Data) {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.E.Renderer.Template(c.W, name, &TemplateData{
		Data: data,
	})
	if err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
	}
}

func New() *Engine {
	engine := &Engine{
		mux: http.NewServeMux(),
	}
	return engine
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	engine.handleRequest(w, r)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) handleRequest(w http.ResponseWriter, r *http.Request) {
	engine.mux.ServeHTTP(w, r)
}

func (engine *Engine) GET(path string, handler func(*Context)) {
	engine.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(&Context{W: w, R: r, E: engine})
	})
}

func (engine *Engine) POST(path string, handler func(*Context)) {
	engine.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(&Context{W: w, R: r, E: engine})
	})
}
