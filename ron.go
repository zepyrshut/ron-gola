package ron

import (
	"context"
	"net/http"
)

type Context struct {
	C context.Context
	W http.ResponseWriter
	R *http.Request
	E *Engine
}

type Engine struct {
	mux *http.ServeMux
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
