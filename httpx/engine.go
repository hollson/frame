package httpx

import (
    "fmt"
    "log"
    "net"
    "sync"
)


// 服务引擎
type Engine struct {
    router *Router  // 路由
    config struct{} // 附：配置
    logger struct{} // 附：日志
}

// 单条请求数据大小为 40k
const MaxRequestSize = 1024 * 40

func Default() *Engine {
    return &Engine{
        router: &Router{
            table: make(routerTable),
            mux:   new(sync.Mutex),
        },
    }
}

func (r *Engine) Run(addr string) error {
    listener, err := net.Listen("tcp4", addr)
    if err != nil {
        return fmt.Errorf("listen error:%v", err)
    }
    log.Println("Http server is running：", listener.Addr().String())
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        go r.handle(conn)
    }
}

func (r *Engine) handle(conn net.Conn) {
    accept := make(chan readerData)
    go r.parseRequest(accept, conn)
    reader := newReader(conn, MaxRequestSize)
    err := reader.read(accept)
    if err != nil {
        fmt.Println(err)
        // 读取数据失败,响应400 bad request错误
        response := newResponse(conn)
        response.errWrite(400)
    }
    close(accept)
    conn.Close()
}

// 监听通道，解析请求数据
func (r *Engine) parseRequest(accept chan readerData, conn net.Conn) {
    for {
        data, isOk := <-accept
        if !isOk {
            return
        }
        request := newContext()
        request.response = newResponse(conn)
        request.parse(data)
        r.handleHTTPRequest(request)
    }
}

// 调用函数处理http请求
func (r *Engine) handleHTTPRequest(request *Context) {
    if f, ok := r.router.table[request.Path]; ok {
        f(request)
        request.response.write()
        log.Printf("\033[0;34m%s\033[0m \033[0;32m OK \033[0m",request.Path)
        return
    }

    // 未找到任何handle 返回404
    request.response.errWrite(404)
}

// 注册路由
func (r *Engine) Register(path string, handle HandlerFunc) {
    r.router.table[path] = handle
}
