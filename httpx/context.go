package httpx

import (
    "context"
    "encoding/json"
    "strings"
)

type H map[string]interface{}

type message struct {
    line   map[string]string // 请求行
    header map[string]string // 请求头
    body   string            // 请求体
}

type Context struct {
    context.Context
    Request *request
    *response
}

func newContext() *Context {
    return &Context{
        Context: context.Background(),
        Request: &request{
            headers: make(map[string]string),
        },
    }
}

// 解析请求内容
func (c *Context) parse(readerData message) {
    c.Request.Method = readerData.line["method"]

    // 解析请求path和get参数
    var queries string
    index := strings.Index(readerData.line["url"], "?")
    if index == -1 {
        c.Request.Path = readerData.line["url"]
    } else {
        c.Request.Path = readerData.line["url"][:index]
        queries = readerData.line["url"][index+1:]
    }
    if c.Request.Method == "GET" {
        // 解析get请求参数
        if queries != "" {
            q := strings.Split(queries, "&")
            for _, v := range q {
                param := strings.Split(v, "=")
                c.Request.queries[param[0]] = param[1]
            }
        }
    } else {
        // 判断content-type类型是不是 application/json
        contentTypes, isExist := c.Request.headers["CONTENT-TYPE"]
        if isExist {
            cTypeArr := strings.Split(contentTypes, ";")
            if strings.EqualFold(cTypeArr[0], "application/json") {
                // 解析post请求参数
                json.Unmarshal([]byte(readerData.body), &(c.Request.posts))
            }
        }
    }
}

// 设置要返回的json数据
func (c *Context) Json(code int, obj interface{}) {
    ret, err := json.Marshal(obj)
    if err == nil {
        // 设置content-length
        c.response.bodyLen = len(ret)
        c.response.body = ret
    }
    c.response.status = code
}

// 设置响应头
func (c *Context) Header(name string, val string) {
    if _, isExist := c.response.headers[name]; !isExist {
        c.response.headers[strings.ToLower(name)] = val
    }
}
