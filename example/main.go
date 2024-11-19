package main

import (
	"log/slog"

	"ron"
)

type SomethingElements struct {
	Name        string
	Description string
}

var elements = []SomethingElements{
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
}

func main() {
	r := ron.New(func(e *ron.Engine) {
		e.LogLevel = slog.LevelDebug
	})

	htmlRender := ron.NewHTMLRender()
	r.Render = htmlRender

	r.Static("static", "static")

	r.GET("/json", helloWorldJSON)
	r.POST("/another", anotherHelloWorld)
	r.GET("/html", helloWorldHTML)
	r.GET("/component", componentHTML)

	r.Run(":8080")
}

func helloWorld(c *ron.Context) {
	slog.Info("Dummy info message")
	c.W.Write([]byte("hello world"))
}

func anotherHelloWorld(c *ron.Context) {
	c.W.Write([]byte("another hello world"))
}

func helloWorldJSON(c *ron.Context) {
	c.JSON(200, ron.Data{"message": "hello world"})
}

func helloWorldHTML(c *ron.Context) {

	pages := ron.Pages{
		TotalElements:   len(elements),
		ElementsPerPage: 5,
	}

	pages.PaginationParams(c.R)
	elementsPaginated := pages.PaginateArray(elements)

	td := &ron.TemplateData{
		Data:  ron.Data{"title": "hello world", "message": "hello world from html", "elements": elementsPaginated},
		Pages: pages,
	}

	c.HTML(200, "page.index.gohtml", td)
}

func componentHTML(c *ron.Context) {
	c.HTML(200, "component.list.gohtml", nil)
}
