package ron

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
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

func Test_findHTMLFiles(t *testing.T) {
	render := defaultHTMLRender()
	if render == nil {
		t.Errorf("Error: %v", render)
		return
	}

	expected := []string{
		"templates\\layout.base.gohtml",
		"templates\\layout.another.gohtml",
		"templates\\fragment.button.gohtml",
		"templates\\component.list.gohtml",
		"templates\\page.index.gohtml",
		"templates\\page.tindex.gohtml",
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
	render := defaultHTMLRender()
	if render == nil {
		t.Errorf("Error: %v", render)
		return
	}

	tc, err := render.createTemplateCache()
	if err != nil || len(tc) != 4 {
		t.Errorf("Error: %v", err)
	}

	templateNames := []string{
		"component.list.gohtml",
		"page.tindex.gohtml",
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
		{"tindex", "<p>layout.base.gohtml</p><p>Foo</p><p>page.tindex.gohtml</p><p>Bar</p>"},
		{"another", "<p>layout.another.gohtml</p><p>Bar</p><p>page.another.gohtml</p><p>Foo</p>"},
	}

	render := defaultHTMLRender()
	if render == nil {
		t.Errorf("Error: %v", render)
		return
	}

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
	var elements []SomethingElements
	for i := 1; i <= 50; i++ {
		elements = append(elements, SomethingElements{
			fmt.Sprintf("element %d", i),
			fmt.Sprintf("description %d", i),
		})
	}
	return elements
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
