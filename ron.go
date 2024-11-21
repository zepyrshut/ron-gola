package ron

import (
	"context"
	"encoding/json"
	"errors"
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

	Middleware func(http.Handler) http.Handler

	responseWriterWrapper struct {
		http.ResponseWriter
		headerWritten bool
	}

	CTX struct {
		W *responseWriterWrapper
		R *http.Request
		E *Engine
	}

	Config struct {
		Timeout  time.Duration
		LogLevel slog.Level
	}

	Engine struct {
		mux        *http.ServeMux
		middleware []Middleware
		groupMux   map[string]*groupMux
		Config     *Config
		Render     *Render
	}

	groupMux struct {
		prefix     string
		mux        *http.ServeMux
		middleware []Middleware
		engine     *Engine
	}
)

const (
	HeaderJSON       string = "application/json"
	HeaderHTML_UTF8  string = "text/html; charset=utf-8"
	HeaderCSS_UTF8   string = "text/css; charset=utf-8"
	HeaderAppJS      string = "application/javascript"
	HeaderPlain_UTF8 string = "text/plain; charset=utf-8"
)

func (w *responseWriterWrapper) WriteHeader(code int) {
	if !w.headerWritten {
		w.headerWritten = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.headerWritten = true
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func defaultEngine() *Engine {
	return &Engine{
		mux:      http.NewServeMux(),
		groupMux: make(map[string]*groupMux),
		Config: &Config{
			Timeout:  time.Second * 30,
			LogLevel: slog.LevelDebug,
		},
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

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = e.mux
	for prefix, group := range e.groupMux {
		if strings.HasPrefix(r.URL.Path, prefix) {
			handler = createStack(group.middleware...)(handler)
			break
		}
	}

	e.middleware = append(e.middleware, e.timeOutMiddleware())
	handler = createStack(e.middleware...)(handler)
	rw := &responseWriterWrapper{ResponseWriter: w}
	handler.ServeHTTP(rw, r)
}

func (e *Engine) Run(addr string) error {
	newLogger(e.Config.LogLevel)
	return http.ListenAndServe(addr, e)
}

func createStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}

func (e *Engine) timeOutMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), e.Config.Timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					slog.Debug("timeout reached")
					http.Error(w, "Request timed out", http.StatusGatewayTimeout)
				}
			case <-done:
			}
		})
	}
}

func (e *Engine) USE(middleware Middleware) {
	e.middleware = append(e.middleware, middleware)
}

func (e *Engine) GET(path string, handler func(*CTX, context.Context)) {
	e.mux.HandleFunc(fmt.Sprintf("GET %s", path), func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w}
		handler(&CTX{W: rw, R: r, E: e}, r.Context())
	})
}

func (e *Engine) POST(path string, handler func(*CTX, context.Context)) {
	e.mux.HandleFunc(fmt.Sprintf("POST %s", path), func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w}
		handler(&CTX{W: rw, R: r, E: e}, r.Context())
	})
}

func (e *Engine) GROUP(prefix string) *groupMux {
	if _, ok := e.groupMux[prefix]; !ok {
		e.groupMux[prefix] = &groupMux{
			prefix: prefix,
			mux:    http.NewServeMux(),
			engine: e,
		}

		e.mux.Handle(prefix+"/", http.StripPrefix(prefix, e.groupMux[prefix].mux))
	}

	return e.groupMux[prefix]
}

func (g *groupMux) USE(middleware Middleware) {
	g.middleware = append(g.middleware, middleware)
}

func (g *groupMux) GET(path string, handler func(*CTX, context.Context)) {
	g.mux.HandleFunc(fmt.Sprintf("GET %s", path), func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w}
		handler(&CTX{W: rw, R: r, E: g.engine}, r.Context())
	})
}

func (g *groupMux) POST(path string, handler func(*CTX, context.Context)) {
	g.mux.HandleFunc(fmt.Sprintf("POST %s", path), func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w}
		handler(&CTX{W: rw, R: r, E: g.engine}, r.Context())
	})
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
func (e *Engine) Static(path, dir string) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	if !strings.HasPrefix(dir, "./") {
		dir = "./" + dir
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Error("static directory does not exist", "path", path, "dir", dir)
		e.mux.Handle(path, http.NotFoundHandler())
		return err
	}

	fs := http.FileServer(http.Dir(dir))
	e.mux.Handle(path, http.StripPrefix(path, fs))
	slog.Info("static files served", "path", path, "dir", dir)
	return nil
}

func (c *CTX) JSON(code int, data any) {
	c.W.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(c.W)
	if err := encoder.Encode(data); err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
		return
	}
	c.W.WriteHeader(code)
}

func (c *CTX) HTML(code int, name string, td *TemplateData) {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.E.Render.Template(c.W, name, td)
	if err != nil {
		http.Error(c.W, err.Error(), http.StatusInternalServerError)
		return
	}
	c.W.WriteHeader(code)
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
