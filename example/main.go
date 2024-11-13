package main

import (
	"log/slog"
	"net/http"

	"ron"
)

func main() {
	r := ron.New()

	htmlRender := ron.NewHTMLRender()
	r.Renderer = htmlRender

	r.GET("/", helloWorld)
	r.GET("/json", helloWorldJSON)
	r.POST("/another", anotherHelloWorld)
	r.GET("/html", helloWorldHTML)
	r.GET("/component", componentHTML)

	slog.Info("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func helloWorld(c *ron.Context) {
	c.W.Write([]byte("hello world"))
}

func anotherHelloWorld(c *ron.Context) {
	c.W.Write([]byte("another hello world"))
}

func helloWorldJSON(c *ron.Context) {
	c.JSON(200, ron.Data{"message": "hello world"})
}

func helloWorldHTML(c *ron.Context) {
	c.HTML(200, "page.index.gohtml", ron.Data{
		"title":   "hello world",
		"message": "hello world from html",
	})
}

func componentHTML(c *ron.Context) {
	c.HTML(200, "component.list.gohtml", ron.Data{})
}
