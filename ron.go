package ron

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
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

	Engine struct {
		mux      *http.ServeMux
		LogLevel slog.Level
		Renderer *Render
	}
)

func DefaultEngine() *Engine {
	return &Engine{
		mux:      http.NewServeMux(),
		LogLevel: slog.LevelInfo,
	}
}

func New(opts ...EngineOptions) *Engine {
	config := DefaultEngine()
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

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handleRequest(w, r)
}

func (e *Engine) Run(addr string) error {
	newLogger(e.LogLevel)
	return http.ListenAndServe(addr, e)
}

func (e *Engine) handleRequest(w http.ResponseWriter, r *http.Request) {
	e.mux.ServeHTTP(w, r)
}

func (e *Engine) GET(path string, handler func(*Context)) {
	e.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(&Context{W: w, R: r, E: e})
	})
}

func (e *Engine) POST(path string, handler func(*Context)) {
	e.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(&Context{W: w, R: r, E: e})
	})
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
