package ron

import (
	"html/template"
	"net/http/httptest"
	"os"
	"reflect"
	"ron/testhelpers"
	"testing"
)

func Test_DefaultHTMLRender(t *testing.T) {
	expected := &Render{
		EnableCache:   false,
		TemplatesPath: "templates",
		Functions:     make(template.FuncMap),
		templateCache: make(templateCache),
	}

	actual := DefaultHTMLRender()
	if reflect.DeepEqual(expected, actual) == false {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}

func Test_HTMLRender(t *testing.T) {
	expected := &Render{
		EnableCache:   false,
		TemplatesPath: "templates",
		Functions:     make(template.FuncMap),
		templateCache: make(templateCache),
	}

	tests := []struct {
		name string
		arg  OptionFunc
	}{
		{"Empty OptionFunc", OptionFunc(func(r *Render) {})},
		{"Nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewHTMLRender(tt.arg)
			if reflect.DeepEqual(expected, actual) == false {
				t.Errorf("Expected: %v, Actual: %v", expected, actual)
			}
		})
	}
}

func Test_apply(t *testing.T) {
	tests := []struct {
		name     string
		expected *Render
		actual   *Render
	}{
		{"Empty OptionFunc", &Render{
			EnableCache:   false,
			TemplatesPath: "templates",
			Functions:     make(template.FuncMap),
			templateCache: make(templateCache),
		}, DefaultHTMLRender()},
		{
			name: "Two OptionFunc", expected: &Render{
				EnableCache:   true,
				TemplatesPath: "foobar",
				Functions:     make(template.FuncMap),
				templateCache: make(templateCache),
			},
			actual: NewHTMLRender(func(r *Render) {
				r.EnableCache = true
				r.TemplatesPath = "foobar"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if reflect.DeepEqual(tt.expected, tt.actual) == false {
				t.Errorf("Expected: %v, Actual: %v", tt.expected, tt.actual)
			}
		})
	}

}

func createDummyFilesAndRender() *Render {
	os.Mkdir("templates", os.ModePerm)

	f, _ := os.Create("templates/layout.base.gohtml")
	f.Write([]byte("{{ define \"layout/base\" }}<p>layout.base.gohtml</p><p>{{ .Data.foo }}</p>{{ block \"base/content\" . }}{{ end }}{{ end }}"))
	f.Close()
	f, _ = os.Create("templates/layout.another.gohtml")
	f.Write([]byte("{{ define \"layout/another\" }}<p>layout.another.gohtml</p><p>{{ .Data.bar }}</p>{{ block \"base/content\" . }}{{ end }}{{ end }}"))
	f.Close()
	f, _ = os.Create("templates/fragment.button.gohtml")
	f.Close()
	f, _ = os.Create("templates/component.list.gohtml")
	f.Close()
	f, _ = os.Create("templates/page.index.gohtml")
	f.Write([]byte("{{ template \"layout/base\" .}}{{ define \"base/content\" }}<p>page.index.gohtml</p><p>{{ .Data.bar }}</p>{{ end }}"))
	f.Close()
	f, _ = os.Create("templates/page.another.gohtml")
	f.Write([]byte("{{ template \"layout/another\" .}}{{ define \"base/content\" }}<p>page.another.gohtml</p><p>{{ .Data.foo }}</p>{{ end }}"))
	f.Close()

	render := DefaultHTMLRender()
	return render
}

func Test_findHTMLFiles(t *testing.T) {
	render := createDummyFilesAndRender()
	if render == nil {
		t.Errorf("Error: %v", render)
		return
	}
	defer os.RemoveAll("templates")

	expected := []string{
		"templates\\layout.base.gohtml",
		"templates\\layout.another.gohtml",
		"templates\\fragment.button.gohtml",
		"templates\\component.list.gohtml",
		"templates\\page.index.gohtml",
		"templates\\page.another.gohtml",
	}
	actual, err := render.findHTMLFiles()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	expectedAny := testhelpers.StringSliceToAnySlice(expected)
	actualAny := testhelpers.StringSliceToAnySlice(actual)

	if testhelpers.CheckSlicesEquality(expectedAny, actualAny) == false {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}

func Test_createTemplateCache(t *testing.T) {
	render := createDummyFilesAndRender()
	if render == nil {
		t.Errorf("Error: %v", render)
	}
	defer os.RemoveAll("templates")

	tc, err := render.createTemplateCache()
	if err != nil || len(tc) != 3 {
		t.Errorf("Error: %v", err)
	}

	templateNames := []string{
		"component.list.gohtml",
		"page.index.gohtml",
		"page.another.gohtml",
	}

	for _, templateName := range templateNames {
		if _, ok := tc[templateName]; ok == false {
			t.Errorf("Error: %v", err)
		}
	}
}

func Test_TemplateDefault(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"index", "<p>layout.base.gohtml</p><p>Foo</p><p>page.index.gohtml</p><p>Bar</p>"},
		{"another", "<p>layout.another.gohtml</p><p>Bar</p><p>page.another.gohtml</p><p>Foo</p>"},
	}

	render := createDummyFilesAndRender()
	if render == nil {
		t.Errorf("Error: %v", render)
	}
	defer os.RemoveAll("templates")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			render.Template(rr, "page."+tt.name+".gohtml", &TemplateData{
				Data: Data{
					"foo": "Foo",
					"bar": "Bar",
				}})

			if rr.Body.String() != tt.expected {
				t.Errorf("Expected: %v, Actual: %v", tt.expected, rr.Body.String())
			}
		})
	}
}
