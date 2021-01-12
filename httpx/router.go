package httpx

type HandlerFunc func(ctx *Context)

type Router map[string]HandlerFunc
