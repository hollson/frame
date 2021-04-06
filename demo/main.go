package main

import (
    "log"

    "frame/httpx"
)

func main() {
    server := httpx.NewServer()
    server.Use(func(c *httpx.Context) {
        log.Printf("\033[0;34m%s\033[0m \033[0;32m OK ~~~~ \033[0m", c.Request.Path)
    })

    server.AddRoute("/get", GetHandler)
    server.AddRoute("/post", PostHandler)
    server.AddRoute("/head", HeadHandler)
    server.Run(":8080")
}

//go:generate curl http://127.0.0.1:8080/get
//go:generate curl --proxy 0.0.0.0:10008 http://172.30.0.86:8080/get
func GetHandler(c *httpx.Context) {
    c.Json(200, httpx.H{
        "hello": "get",
    })
}

//go:generate curl http://127.0.0.1:8080/post
//go:generate curl --proxy 0.0.0.0:10008 http://127.0.0.1:8080/post
func PostHandler(c *httpx.Context) {
    c.Json(200, httpx.H{
        "hello": "post",
    })
}

//go:generate curl -v http://127.0.0.1:8080/head
//go:generate curl --proxy 0.0.0.0:10008 http://172.30.0.86:8080/head
func HeadHandler(c *httpx.Context) {
    c.Header("redirect", "www.abc.com")
}
