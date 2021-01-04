package httpx

import (
    "fmt"
    "net"
)

// 单条请求数据大小为 40k
const MaxRequestSize = 1024 * 40

type Server interface {
    AddRoute(path string, handle HandlerFunc)
    Run(addr string) error
}

// 服务引擎
type server struct {
    router     *Router       // 路由
    middleware []HandlerFunc // 中间件
    // logger     log.Logger      // 日志
}

// 创建服务实例
func NewServer() *server {
    return &server{
        router:     &Router{make(map[string]HandlerFunc)},
        middleware: make([]HandlerFunc, 0),
    }
}

// 注册路由
func (r *server) AddRoute(path string, handle HandlerFunc) {
    r.router.table[path] = handle
}

// 中间件
func (r *server) Use(f HandlerFunc) {
    r.middleware = append(r.middleware, f)
}

// 运行服务
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

    // 处理请求
    go r.requestHandler(msg, conn)

    // 处理响应
    reader := newReader(conn, MaxRequestSize)
    err := reader.responseHandler(msg) // 非阻塞
    if err != nil {
        response := newResponse(conn)
        response.errWrite(400)
    }

}

// 监听通道，解析请求数据
func (r *server) requestHandler(accept chan message, conn net.Conn) {
    for {
        data, isOk := <-accept
        if !isOk {
            return
        }
        req := newContext()
        req.response = newResponse(conn)
        req.parse(data)

        // 执行中间件
        for _, f := range r.middleware {
            f(req)
        }
        if f, ok := r.router.table[req.Request.Path]; ok {

            f(req)
            req.response.write()
            return
        }

        req.response.errWrite(404)
    }
}
