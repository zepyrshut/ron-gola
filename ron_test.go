package ron

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_defaultEngine(t *testing.T) {
	e := defaultEngine()
	if e == nil {
		t.Error("Expected engine, Actual: nil")
	}
}

func Test_New(t *testing.T) {
	e := New()
	if e == nil {
		t.Error("Expected engine, Actual: nil")
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
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusNotFound, status)
	}
}

func Test_GET(t *testing.T) {
	e := New()
	e.GET("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "GET" {
		t.Errorf("Expected: GET, Actual: %s", rr.Body.String())
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

func Test_USE(t *testing.T) {
	e := New()
	e.USE(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("MIDDLEWARE"))
			next.ServeHTTP(w, r)
		})
	})
	e.GET("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "MIDDLEWAREGET" {
		t.Errorf("Expected: MIDDLEWAREGET, Actual: %s", rr.Body.String())
	}
}

func Test_GROUP(t *testing.T) {
	tests := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{"GET", "/group/", http.StatusOK, "GET"},
		{"POST", "/group/", http.StatusOK, "POST"},
	}

	e := New()
	g := e.GROUP("/group")
	g.GET("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET"))
	})
	g.POST("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("POST"))
	})

	for _, tt := range tests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		e.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.expectedCode {
			t.Errorf("Expected status code: %d, Actual: %d", tt.expectedCode, status)
		}

		if rr.Body.String() != tt.expectedBody {
			t.Errorf("Expected: %s, Actual: %s", tt.expectedBody, rr.Body.String())
		}
	}
}

func Test_GROUPUSE(t *testing.T) {
	e := New()
	g := e.GROUP("/group")
	g.USE(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("MIDDLEWARE"))
			next.ServeHTTP(w, r)
		})
	})
	g.GET("/", func(c *Context) {
		c.W.WriteHeader(http.StatusOK)
		c.W.Write([]byte("GET"))
	})

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/group/", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Body.String() != "MIDDLEWAREGET" {
		t.Errorf("Expected: MIDDLEWAREGET, Actual: %s", rr.Body.String())
	}
}

func Test_Static(t *testing.T) {
	os.Mkdir("assets", os.ModePerm)
	f, _ := os.Create("assets/style.css")
	f.WriteString("body { background-color: red; }")
	f.Close()
	defer os.Remove("assets/style.css")

	e := New()
	e.Static("assets", "assets")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/style.css", nil)
	e.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}
}

type Foo struct {
	Bar string  `json:"bar"`
	Taz int     `json:"something"`
	Car *string `json:"car"`
}

func Test_JSON(t *testing.T) {
	rr := httptest.NewRecorder()
	c := &Context{
		W: rr,
	}

	expected := `{"bar":"bar","something":30,"car":null}`

	c.JSON(http.StatusOK, Foo{
		Bar: "bar",
		Taz: 30,
		Car: nil,
	})

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type: application/json, Actual: %s", c.W.Header().Get("Content-Type"))
	}

	if rr.Body.String() != string(expected)+"\n" {
		t.Errorf("Expected: %s, Actual: %s", string(expected), rr.Body.String())
	}
}

func Test_HTML(t *testing.T) {
	os.Mkdir("templates", os.ModePerm)
	f, _ := os.Create("templates/page.index.gohtml")
	f.WriteString("<h1>{{.Data.heading1}}</h1><h2>{{.Data.heading2}}</h2>")
	f.Close()
	defer os.RemoveAll("templates")

	rr := httptest.NewRecorder()
	c := &Context{
		W: rr,
		E: &Engine{
			Render: NewHTMLRender(),
		},
	}

	expected := `<h1>foo</h1><h2>bar</h2>`

	c.HTML(http.StatusOK, "page.index.gohtml", &TemplateData{
		Data: Data{
			"heading1": "foo",
			"heading2": "bar",
		},
	})

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code: %d, Actual: %d", http.StatusOK, status)
	}

	if rr.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type: text/html; charset=utf-8, Actual: %s", c.W.Header().Get("Content-Type"))
	}

	if rr.Body.String() != string(expected) {
		t.Errorf("Expected: %s, Actual: %s", string(expected), rr.Body.String())
	}
}
