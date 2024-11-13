package ron

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

type (
	templateCache map[string]*template.Template

	TemplateData struct {
		Data Data
	}

	OptionFunc func(*Render)
	Render     struct {
		EnableCache   bool
		TemplatesPath string
		Functions     template.FuncMap
		templateCache templateCache
	}
)

func DefaultHTMLRender() *Render {
	return &Render{
		EnableCache:   false,
		TemplatesPath: "templates",
		Functions:     make(template.FuncMap),
		templateCache: make(templateCache),
	}
}

func NewHTMLRender(opts ...OptionFunc) *Render {
	config := DefaultHTMLRender()
	return config.apply(opts...)
}

func (re *Render) apply(opts ...OptionFunc) *Render {
	for _, opt := range opts {
		if opt != nil {
			opt(re)
		}
	}

	return re
}

func (re *Render) Template(w http.ResponseWriter, tmpl string, td *TemplateData) error {
	var tc templateCache
	var err error

	if td == nil {
		td = &TemplateData{}
	}

	if re.EnableCache {
		tc = re.templateCache
	} else {
		tc, err = re.createTemplateCache()
		if err != nil {
			return err
		}
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, td)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
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
