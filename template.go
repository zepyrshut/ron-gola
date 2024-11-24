package ron

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type (
	templateCache map[string]*template.Template

	TemplateData struct {
		Data  Data
		Pages Pages
	}

	RenderOptions func(*Render)
	Render        struct {
		EnableCache   bool
		TemplatesPath string
		Functions     template.FuncMap
		TemplateData  TemplateData
		templateCache templateCache
	}
)

func defaultHTMLRender() *Render {
	return &Render{
		EnableCache:   false,
		TemplatesPath: "templates",
		TemplateData:  TemplateData{},
		Functions:     template.FuncMap{},
		templateCache: templateCache{},
	}
}

func NewHTMLRender(opts ...RenderOptions) *Render {
	config := defaultHTMLRender()
	return config.apply(opts...)
}

func (re *Render) apply(opts ...RenderOptions) *Render {
	for _, opt := range opts {
		if opt != nil {
			opt(re)
		}
	}

	return re
}

func defaultIfEmpty(fallback, value string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (re *Render) Template(w http.ResponseWriter, tmpl string, td *TemplateData) error {
	var tc templateCache
	var err error

	re.Functions["default"] = defaultIfEmpty

	if td == nil {
		td = &TemplateData{}
	}

	tc, err = re.getTemplateCache()
	if err != nil {
		return err
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, td); err != nil {
		return err
	}

	if _, err = buf.WriteTo(w); err != nil {
		return err
	}

	return nil
}

func (re *Render) getTemplateCache() (templateCache, error) {
	slog.Debug("template cache", "tc status", re.EnableCache, "tc", len(re.templateCache))
	if len(re.templateCache) == 0 {
		cachedTemplates, err := re.createTemplateCache()
		if err != nil {
			return nil, err
		}
		re.templateCache = cachedTemplates
	}
	if re.EnableCache {
		return re.templateCache, nil
	}
	return re.createTemplateCache()
}

func (re *Render) findHTMLFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(re.TemplatesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".gohtml" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (re *Render) createTemplateCache() (templateCache, error) {
	cache := templateCache{}
	var baseTemplates []string
	var renderTemplates []string

	templates, err := re.findHTMLFiles()
	if err != nil {
		return cache, err
	}

	for _, file := range templates {
		filePathBase := filepath.Base(file)
		if strings.Contains(filePathBase, "layout") || strings.Contains(filePathBase, "fragment") {
			baseTemplates = append(baseTemplates, file)
		}
	}

	for _, file := range templates {
		filePathBase := filepath.Base(file)
		if strings.Contains(filePathBase, "page") || strings.Contains(filePathBase, "component") {
			renderTemplates = append(baseTemplates, file)
			ts, err := template.New(filePathBase).Funcs(re.Functions).ParseFiles(append(baseTemplates, renderTemplates...)...)
			if err != nil {
				return cache, err
			}
			cache[filePathBase] = ts
		}
	}

	return cache, nil
}

// Pages contains pagination info.
type Pages struct {
	// TotalElements indicates the total number of elements available for
	// pagination.
	TotalElements int
	// ElementsPerPage defines the number of elements to display per page in
	// pagination.
	ElementsPerPage int
	// ActualPage represents the current page number in pagination.
	ActualPage int
}

func (p *Pages) PaginationParams(r *http.Request) {
	limit := r.FormValue("limit")
	page := r.FormValue("page")

	if limit == "" {
		if p.ElementsPerPage != 0 {
			limit = strconv.Itoa(p.ElementsPerPage)
		} else {
			limit = "20"
		}
	}

	if page == "" || page == "0" {
		if p.ActualPage != 0 {
			page = strconv.Itoa(p.ActualPage)
		} else {
			page = "1"
		}
	}

	limitInt, _ := strconv.Atoi(limit)
	pageInt, _ := strconv.Atoi(page)
	offset := (pageInt - 1) * limitInt
	currentPage := offset/limitInt + 1

	p.ElementsPerPage = limitInt
	p.ActualPage = currentPage
}

func (p Pages) PaginateArray(elements any) any {
	itemsValue := reflect.ValueOf(elements)

	if p.ActualPage < 1 {
		p.ActualPage = 1
	}

	if p.ActualPage > p.TotalPages() {
		p.ActualPage = p.TotalPages()
	}

	startIndex := (p.ActualPage - 1) * p.ElementsPerPage
	endIndex := startIndex + p.ElementsPerPage

	return itemsValue.Slice(startIndex, endIndex).Interface()
}

func (p Pages) CurrentPage() int {
	return p.ActualPage
}

func (p Pages) TotalPages() int {
	return (p.TotalElements + p.ElementsPerPage - 1) / p.ElementsPerPage
}

func (p Pages) IsFirst() bool {
	return p.ActualPage == 1
}

func (p Pages) IsLast() bool {
	return p.ActualPage == p.TotalPages()
}

func (p Pages) HasPrevious() bool {
	return p.ActualPage > 1
}

func (p Pages) HasNext() bool {
	return p.ActualPage < p.TotalPages()
}

func (p Pages) Previous() int {
	if p.ActualPage > p.TotalPages() {
		return p.TotalPages()
	}
	return p.ActualPage - 1
}

func (p Pages) Next() int {
	if p.ActualPage < 1 {
		return 1
	}
	return p.ActualPage + 1
}

func (p Pages) GoToPage(page int) int {
	if page < 1 {
		page = 1
	} else if page > p.TotalPages() {
		page = p.TotalPages()
	}
	return page
}

func (p Pages) First() int {
	return p.GoToPage(1)
}

func (p Pages) Last() int {
	return p.GoToPage(p.TotalPages())
}

// Page contiene la información de una página. Utilizado para la barra de
// paginación que suelen mostrarse en la parte inferior de una lista o tabla.

// Page represents a single page in pagination, including its number and active
// state. Useful for pagination bar.
type Page struct {
	// Number is the numeric identifier of the page in pagination.
	Number int
	// Active indicates if the page is the currently selected page.
	Active bool
}

func (p Page) NumberOfPage() int {
	return p.Number
}

func (p Page) IsActive() bool {
	return p.Active
}

// PageRange generates a slice of Page instances representing a range of pages
// to be displayed in a pagination bar.
func (p Pages) PageRange(maxPagesToShow int) []Page {
	var pages []Page
	totalPages := p.TotalPages()

	startPage := p.ActualPage - (maxPagesToShow / 2)
	endPage := p.ActualPage + (maxPagesToShow / 2)

	if startPage < 1 {
		startPage = 1
		endPage = maxPagesToShow
	}

	if endPage > totalPages {
		endPage = totalPages
		startPage = totalPages - maxPagesToShow + 1
		if startPage < 1 {
			startPage = 1
		}
	}

	for i := startPage; i <= endPage; i++ {
		pages = append(pages, Page{
			Number: i,
			Active: i == p.ActualPage,
		})
	}

	return pages
}
