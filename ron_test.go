package ron

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"ron/testhelpers"
	"testing"
)

func TestMain(m *testing.M) {
	os.Mkdir("templates", os.ModePerm)
	f, _ := os.Create("templates/page.index.gohtml")
	f.WriteString("<h1>{{.Data.heading1}}</h1><h2>{{.Data.heading2}}</h2>")
	f.Close()

	os.Mkdir("assets", os.ModePerm)
	f, _ = os.Create("assets/style.css")
	f.WriteString("body { background-color: red; }")
	f.Close()

	f, _ = os.Create("assets/script.js")
	f.WriteString("console.log('Hello, World!');")
	f.Close()

	code := m.Run()

	os.RemoveAll("templates")
	os.RemoveAll("assets")

	os.Exit(code)
}

func Test_defaultEngine(t *testing.T) {
	e := defaultEngine()
	if e == nil {
		t.Error("Expected Engine, Actual: nil")
	}
}

func Test_New(t *testing.T) {
	e := New()
	if e == nil {
		t.Error("Expected Engine, Actual: nil")
	}
	if e.Render != nil {
		t.Error("No expected Renderer, Actual: Renderer")
	}
}

func Test_applyEngineConfig(t *testing.T) {
	e := New(func(e *Engine) {
		e.Render = NewHTMLRender()
		e.LogLevel = 1
	})
	if e.Render == nil {
		t.Error("Expected Renderer, Actual: nil")
	}
	if e.LogLevel != 1 {
		t.Errorf("Expected LogLevel: 1, Actual: %d", e.LogLevel)
	}
}

func Test_ServeHTTP(t *testing.T) {
	e := New()
	api := e.GROUP("/api")
	api.GET("/index", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET API"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/index", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}
}

func Test_RUN(t *testing.T) {
	e := New()
	go func() {
		e.Run(":8080")
	}()
}

func Test_createStack(t *testing.T) {
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 1"))
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 2"))
			next.ServeHTTP(w, r)
		})
	}

	stack := createStack(m1, m2)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	stack(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Handler"))
	})).ServeHTTP(rr, req)

	if rr.Body.String() != "Middleware 1Middleware 2Handler" {
		t.Errorf("Expected: Middleware 1Middleware 2Handler, Actual: %s", rr.Body.String())
	}
}

func Test_USE(t *testing.T) {
	e := New()
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 1"))
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 2"))
			next.ServeHTTP(w, r)
		})
	}

	e.USE(m1)
	e.USE(m2)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(rr, req)

	if rr.Body.String() != "Middleware 1Middleware 2404 page not found\n" {
		t.Errorf("Expected: Middleware 1Middleware 2Handler, Actual: %s", rr.Body.String())
	}
}

func Test_GET(t *testing.T) {
	e := New()
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"root endpoint", "GET", "/", http.StatusOK, "GET Root"},
		{"api endpoint", "GET", "/api", http.StatusOK, "GET API"},
		{"api endpoint with version", "GET", "/api/v1", http.StatusOK, "GET API v1"},
		{"resource with param", "GET", "/api/v1/resource/1", http.StatusOK, "GET Resource"},
	}

	e.GET("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET Root"))
	})
	e.GET("/api", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET API"))
	})
	e.GET("/api/v1", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET API v1"))
	})
	e.GET("/api/v1/resource/{id}", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET Resource"))
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			e.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code: %d, Actual: %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body: %q, Actual: %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func Test_POST(t *testing.T) {
	e := New()
	e.POST("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("POST"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "POST" {
		t.Errorf("Expected: POST, Actual: %s", rr.Body.String())
	}
}

func Test_GROUP(t *testing.T) {
	e := New()
	api := e.GROUP("/api")
	api.GET("/index", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET API"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/index", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "GET API" {
		t.Errorf("Expected: GET API, Actual: %s", rr.Body.String())
	}
}

func Test_GROUPWithMiddleware(t *testing.T) {
	e := New()
	e.GET("/index", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET Root"))
	})
	e.USE(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 1"))
			next.ServeHTTP(w, r)
		})
	})

	api := e.GROUP("/api")
	api.USE(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Middleware 2"))
			next.ServeHTTP(w, r)
		})
	})
	api.GET("/index", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET API"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/index", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "Middleware 1Middleware 2GET API" {
		t.Errorf("Expected: Middleware 1Middleware 2GET API, Actual: %s", rr.Body.String())
	}
}

func Test_GROUPPOST(t *testing.T) {
	e := New()
	api := e.GROUP("/api")
	api.POST("/index", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("POST API"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/index", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "POST API" {
		t.Errorf("Expected: POST API, Actual: %s", rr.Body.String())
	}
}

func Test_Static(t *testing.T) {
	tests := map[string]struct {
		givenPath        string
		givenFile        string
		givenDirectory   string
		expectedResponse testhelpers.ExpectedResponse
	}{
		"valid css path": {
			givenPath:      "assets",
			givenFile:      "style.css",
			givenDirectory: "assets",
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusOK,
				Header: HeaderCSS_UTF8,
				Body:   "body { background-color: red; }",
			},
		},
		"valid js path": {
			givenPath:      "assets",
			givenFile:      "script.js",
			givenDirectory: "assets",
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusOK,
				Header: HeaderJS_UTF8,
				Body:   "console.log('Hello, World!');",
			},
		},
		"invalid path": {
			givenPath:      "assets",
			givenFile:      "nonexistent.css",
			givenDirectory: "assets",
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusNotFound,
				Header: HeaderPlain_UTF8,
				Body:   "404 page not found\n",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			e := New()
			e.Static(tt.givenPath, tt.givenDirectory)

			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/%s", tt.givenPath, tt.givenFile), nil)
			e.ServeHTTP(rr, req)

			testhelpers.VerifyResponse(t, rr, tt.expectedResponse)
		})
	}
}

type Foo struct {
	Bar string  `json:"bar"`
	Taz int     `json:"something"`
	Car *string `json:"car"`
}

func Test_JSON(t *testing.T) {
	tests := map[string]struct {
		givenCode        int
		givenData        any
		expectedResponse testhelpers.ExpectedResponse
	}{
		"valid JSON": {
			givenCode: http.StatusOK,
			givenData: Foo{Bar: "bar", Taz: 30, Car: nil},
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusOK,
				Header: HeaderJSON,
				Body:   `{"bar":"bar","something":30,"car":null}` + "\n",
			},
		},
		"invalid JSON": {
			givenCode: http.StatusOK,
			givenData: make(chan int),
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusInternalServerError,
				Header: HeaderPlain_UTF8,
				Body:   "json: unsupported type: chan int\n",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			c := &Context{
				W: rr,
			}

			c.JSON(tt.givenCode, tt.givenData)

			testhelpers.VerifyResponse(t, rr, tt.expectedResponse)
		})
	}
}

func Test_HTML(t *testing.T) {
	tests := map[string]struct {
		givenCode        int
		givenTemplate    string
		givenData        *TemplateData
		expectedResponse testhelpers.ExpectedResponse
	}{
		"valid HTML": {
			givenCode:     http.StatusOK,
			givenTemplate: "page.index.gohtml",
			givenData:     &TemplateData{Data: Data{"heading1": "foo", "heading2": "bar"}},
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusOK,
				Header: HeaderHTML_UTF8,
				Body:   "<h1>foo</h1><h2>bar</h2>",
			},
		},
		"template not found": {
			givenCode:     http.StatusOK,
			givenTemplate: "nonexistent.gohtml",
			givenData:     &TemplateData{Data: Data{"heading1": "foo", "heading2": "bar"}},
			expectedResponse: testhelpers.ExpectedResponse{
				Code:   http.StatusInternalServerError,
				Header: HeaderPlain_UTF8,
				Body:   "can't get template from cache\n",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			c := &Context{
				W: rr,
				E: &Engine{
					Render: NewHTMLRender(),
				},
			}

			c.HTML(tt.givenCode, tt.givenTemplate, tt.givenData)

			testhelpers.VerifyResponse(t, rr, tt.expectedResponse)
		})
	}

}

func Test_newLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   slog.Level
		wantErr bool
	}{
		{"valid level", 1, false},
		{"invalid level", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("newLogger() panicked: %v", r)
					}
				}
			}()
			newLogger(tt.level)
		})
	}
}
