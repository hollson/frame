package httpx

import (
	"fmt"
	"log"
	"net"
	"sync"

)

// 服务引擎
type Server struct {
	router *Router  // 路由
	config struct{} // 附：配置
	logger struct{} // 附：日志
}

// 单条请求数据大小为 40k
const MaxRequestSize = 1024 * 40

// 创建服务实例
func NewServer() *Server {
	return &Server{
		router: &Router{
			table: make(routerTable),
			mux:   new(sync.Mutex),
		},
	}
}

// 注册路由
func (r *Server) AddRoute(path string, handle HandlerFunc) {
	r.router.table[path] = handle
}

// 运行服务
func (r *Server) Run(addr string) error {
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

// 处理程序
func (r *Server) process(conn net.Conn) {

	msg := make(chan readerData)
	go r.parseRequest(msg, conn)

	reader := newReader(conn, MaxRequestSize)
	err := reader.read(msg) // 非阻塞
	if err != nil {
		fmt.Println(err)
		// 读取数据失败,响应400 bad request错误
		response := newResponse(conn)
		response.errWrite(400)
	}
	close(msg)
	conn.Close()
}

// 监听通道，解析请求数据
func (r *Server) parseRequest(accept chan readerData, conn net.Conn) {
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
func (r *Server) handleHTTPRequest(request *Context) {
	if f, ok := r.router.table[request.Path]; ok {
		f(request)
		request.response.write()
		log.Printf("\033[0;34m%s\033[0m \033[0;32m OK \033[0m", request.Path)
		return
	}

	// 未找到任何handle 返回404
	request.response.errWrite(404)
}
