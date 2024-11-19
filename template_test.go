package ron

import (
	"html/template"
	"net/http"
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
		TemplateData:  TemplateData{},
		Functions:     template.FuncMap{},
		templateCache: templateCache{},
	}

	actual := defaultHTMLRender()
	if reflect.DeepEqual(expected, actual) == false {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}

func Test_HTMLRender(t *testing.T) {
	expected := &Render{
		EnableCache:   false,
		TemplatesPath: "templates",
		TemplateData:  TemplateData{},
		Functions:     template.FuncMap{},
		templateCache: templateCache{},
	}

	tests := []struct {
		name string
		arg  RenderOptions
	}{
		{"Empty OptionFunc", RenderOptions(func(r *Render) {})},
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

func Test_applyRenderConfig(t *testing.T) {
	tests := []struct {
		name     string
		expected *Render
		actual   *Render
	}{
		{"Empty OptionFunc", &Render{
			EnableCache:   false,
			TemplatesPath: "templates",
			TemplateData:  TemplateData{},
			Functions:     template.FuncMap{},
			templateCache: templateCache{},
		}, defaultHTMLRender()},
		{
			name: "Two OptionFunc", expected: &Render{
				EnableCache:   true,
				TemplatesPath: "foobar",
				TemplateData:  TemplateData{},
				Functions:     template.FuncMap{},
				templateCache: templateCache{},
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

	render := defaultHTMLRender()
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

type SomethingElements struct {
	Name        string
	Description string
}

func createDummyElements() []SomethingElements {
	return []SomethingElements{
		{"element 1", "description 1"},
		{"element 2", "description 2"},
		{"element 3", "description 3"},
		{"element 4", "description 4"},
		{"element 5", "description 5"},
		{"element 6", "description 6"},
		{"element 7", "description 7"},
		{"element 8", "description 8"},
		{"element 9", "description 9"},
		{"element 10", "description 10"},
		{"element 11", "description 11"},
		{"element 12", "description 12"},
		{"element 13", "description 13"},
		{"element 14", "description 14"},
		{"element 15", "description 15"},
		{"element 16", "description 16"},
		{"element 17", "description 17"},
		{"element 18", "description 18"},
		{"element 19", "description 19"},
		{"element 20", "description 20"},
		{"element 21", "description 21"},
		{"element 22", "description 22"},
		{"element 23", "description 23"},
		{"element 24", "description 24"},
		{"element 25", "description 25"},
		{"element 26", "description 26"},
		{"element 27", "description 27"},
		{"element 28", "description 28"},
		{"element 29", "description 29"},
		{"element 30", "description 30"},
		{"element 31", "description 31"},
		{"element 32", "description 32"},
		{"element 33", "description 33"},
		{"element 34", "description 34"},
		{"element 35", "description 35"},
		{"element 36", "description 36"},
		{"element 37", "description 37"},
		{"element 38", "description 38"},
		{"element 39", "description 39"},
		{"element 40", "description 40"},
		{"element 41", "description 41"},
		{"element 42", "description 42"},
		{"element 43", "description 43"},
		{"element 44", "description 44"},
		{"element 45", "description 45"},
		{"element 46", "description 46"},
		{"element 47", "description 47"},
		{"element 48", "description 48"},
		{"element 49", "description 49"},
		{"element 50", "description 50"},
	}
}

func Test_PaginationParams(t *testing.T) {
	elements := createDummyElements()
	tests := []struct {
		name     string
		req      *http.Request
		given    *Pages
		expected *Pages
	}{
		{
			name: "All defaults",
			req:  httptest.NewRequest("GET", "/", nil),
			given: &Pages{
				TotalElements: len(elements),
			},
			expected: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 20,
				ActualPage:      1,
			},
		},
		{
			name: "Without params",
			req:  httptest.NewRequest("GET", "/", nil),
			given: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      3,
			},
			expected: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      3,
			},
		},
		{
			name: "With page param",
			req:  httptest.NewRequest("GET", "/?page=2", nil),
			given: &Pages{
				TotalElements: len(elements),
			},
			expected: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 20,
				ActualPage:      2,
			},
		},
		{
			name: "With limit param",
			req:  httptest.NewRequest("GET", "/?limit=10", nil),
			given: &Pages{
				TotalElements: len(elements),
			},
			expected: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      1,
			},
		},
		{
			name: "With page and limit param",
			req:  httptest.NewRequest("GET", "/?page=2&limit=10", nil),
			given: &Pages{
				TotalElements: len(elements),
			},
			expected: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.given.PaginationParams(tt.req)
			if !reflect.DeepEqual(tt.expected, tt.given) {
				t.Errorf("Expected: %v, Actual: %v", tt.expected, tt.given)
			}
		})
	}
}

func Test_PaginateArray(t *testing.T) {
	elements := createDummyElements()
	tests := []struct {
		name     string
		given    *Pages
		expected []SomethingElements
	}{
		{
			name: "First page",
			given: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      1,
			},
			expected: elements[:10],
		},
		{
			name: "Second page",
			given: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      2,
			},
			expected: elements[10:20],
		},
		{
			name: "Out of range superior",
			given: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      999,
			},
			expected: elements[40:50],
		},
		{
			name: "Out of range inferior",
			given: &Pages{
				TotalElements:   len(elements),
				ElementsPerPage: 10,
				ActualPage:      -1,
			},
			expected: elements[:10],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.given.PaginateArray(elements)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("Expected: %v, Actual: %v", tt.expected, actual)
			}
		})
	}
}
