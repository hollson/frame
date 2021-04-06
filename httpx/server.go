package httpx

import (
    "fmt"
    "net"
    "time"
)

const MaxRequestSize = (1 << 10) * 4 // 4K

// 一个web-server，本质是一个路由器对象，具备的要点：
//  1. 持有一个路由集合，以处理多路服务端请求；
//  2. 一个TCP的端口监听服务，用于接收请求与响应请求；
//  其次，可以附加中间件等功能。
type Server interface {
    AddRoute(path string, f HandlerFunc)
    Use(f HandlerFunc)
    Run(addr string) error
}

// 服务引擎(只包含两个路由集合字段)
type server struct {
    router     Router        // 路由表
    middleware []HandlerFunc // 中间件
}

func NewServer() *server {
    return &server{
        router:     make(map[string]HandlerFunc),
        middleware: make([]HandlerFunc, 0),
    }
}

// 向路由表添加路由信息
func (r *server) AddRoute(path string, f HandlerFunc) {
    r.router[path] = f
}

// 添加中间件
func (r *server) Use(f HandlerFunc) {
    r.middleware = append(r.middleware, f)
}

// 启动一个TCP监听服务
func (r *server) Run(addr string) error {
    listener, err := net.Listen("tcp4", addr)
    if err != nil {
        return fmt.Errorf("%v", err)
    }
    fmt.Println("服务已启动：", listener.Addr().String())

    // 阻塞式服务
    for {
        conn, err := listener.Accept() // 接受多个客户端连接
        if err != nil {
            fmt.Println(err)
            continue
        }
        go r.process(conn)
    }
}

// 处理请求和响应
func (r *server) process(conn net.Conn) {
    msg := make(chan message)
    defer conn.Close()
    defer close(msg)

    // 处理响应(监听报文)
    go r.responseHandler(msg, conn)

    // 处理请求
    reader := newReader(conn, MaxRequestSize)
    err := reader.requestHandler(msg)
    if err != nil {
        response := newResponse(conn)
        response.writeErr(400)
    }
}

// 处理上下文
func (r *server) responseHandler(msg chan message, c net.Conn) {
    var ctx = newContext()
    ctx.response = newResponse(c)
    for {
        data, ok := <-msg
        if !ok {
            time.Sleep(time.Second)
            return
        }
        ctx.parse(data)

        // 执行中间件
        for _, f := range r.middleware {
            f(ctx)
        }

        // 处理请求
        if f, ok := r.router[ctx.Request.Path]; ok {
            f(ctx)
            ctx.response.write() // fixme：输出
            return
        }
        ctx.response.writeErr(404)
    }
}
