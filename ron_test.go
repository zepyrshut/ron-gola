package ron

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type Foo struct {
	Bar string  `json:"bar"`
	Taz int     `json:"taz"`
	Car *string `json:"car"`
}

func Test_JSON(t *testing.T) {
	rr := httptest.NewRecorder()
	c := &Context{
		W: rr,
	}

	expected := `{"bar":"bar","taz":30,"car":null}`

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
			Renderer: NewHTMLRender(),
		},
	}

	expected := `<h1>foo</h1><h2>bar</h2>`

	c.HTML(http.StatusOK, "page.index.gohtml", Data{
		"heading1": "foo",
		"heading2": "bar",
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
