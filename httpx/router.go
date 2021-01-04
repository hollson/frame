package httpx

type HandlerFunc func(ctx *Context)

type table map[string]HandlerFunc

// 子路由配置
type Router struct {
    table         // 路由表

}
