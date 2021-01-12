package httpx

import (
    "fmt"
    "net"
    "time"
)

const MaxRequestSize = (1 << 10) * 4 // 4K

type Server interface {
    AddRoute(path string, f HandlerFunc)
    Use(f HandlerFunc)
    Run(addr string) error
}

// 服务引擎
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

func (r *server) AddRoute(path string, f HandlerFunc) {
    r.router[path] = f
}

func (r *server) Use(f HandlerFunc) {
    r.middleware = append(r.middleware, f)
}

func (r *server) Run(addr string) error {
    listener, err := net.Listen("tcp4", addr)
    if err != nil {
        return fmt.Errorf("%v", err)
    }

    fmt.Println("服务已启动：", listener.Addr().String())
    for {
        // 阻塞等待连接
        conn, err := listener.Accept()
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
