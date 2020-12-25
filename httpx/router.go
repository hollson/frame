package httpx

import (
    "container/list"
    "sync"
)

type HandlerFunc func(ctx *Context)

type routerTable map[string]HandlerFunc

// 子路由配置
type Router struct {
    table      routerTable // 路由表
    // isRoot     bool
    mux        *sync.Mutex
    Handler    HandlerFunc   // 处理程序
    // Middleware []HandlerFunc // 中间件
    Method     string        // 定义的请求的方法
    // fullPath   string        // 定义的完整路径
    // server      *HttpServer //全局的http服务入口，有rider传入
    handlerList *list.List // 中间件加处理函数链表（会存放在context中)
}
