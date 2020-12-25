package main

import (
    "fmt"
    "frame/httpx"
)

//go:generate curl http://127.0.0.1:8080/get
func main() {
    router := httpx.Default()
    router.Register("/get", GetHandler)
    router.Register("/post", PostHandler)
    router.Run(":8080")
}

func GetHandler(c *httpx.Context) {
    c.Json(200, httpx.H{
        "hello": "world",
    })
    fmt.Println()
}

func PostHandler(c *httpx.Context) {
    c.Json(200, httpx.H{
        "hello": "world",
    })
    fmt.Println()
}