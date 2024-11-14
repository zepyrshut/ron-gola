package main

import (
	"log/slog"

	"ron"
)

func main() {
	r := ron.New(func(e *ron.Engine) {
		e.LogLevel = slog.LevelDebug
	})

	htmlRender := ron.NewHTMLRender()
	r.Renderer = htmlRender

	r.GET("/", helloWorld)
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
	c.HTML(200, "page.index.gohtml", ron.Data{
		"title":   "hello world",
		"message": "hello world from html",
	})
}

func componentHTML(c *ron.Context) {
	c.HTML(200, "component.list.gohtml", ron.Data{})
}
