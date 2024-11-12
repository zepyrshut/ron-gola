package main

import (
	"log/slog"
	"net/http"

	"ron"
)

func main() {
	r := ron.New()

	r.GET("/", helloWorld)
	r.POST("/another", anotherHelloWorld)

	slog.Info("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func helloWorld(c *ron.Context) {
	c.W.Write([]byte("hello world"))
}

func anotherHelloWorld(c *ron.Context) {
	c.W.Write([]byte("another hello world"))
}
