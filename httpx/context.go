package httpx

import (
    "context"
    "encoding/json"
    "strings"
)

type H map[string]interface{}

type request struct {
    Path    string
    Method  string
    headers map[string]string
    queries map[string]string
    posts   map[string]string
}

type Context struct {
    context.Context
    *request
    *response
}

func newContext() *Context {
    return &Context{
        Context: context.Background(),
        request: &request{
            headers: make(map[string]string),
            queries: make(map[string]string),
            posts:   make(map[string]string),
        },
    }
}

// 解析请求内容
func (c *Context) parse(readerData readerData) {
    c.Method = readerData.line["method"]
    c.request.headers = readerData.header

    // 解析请求path和get参数
    var queries string
    index := strings.Index(readerData.line["url"], "?")
    if index == -1 {
        c.Path = readerData.line["url"]
    } else {
        c.Path = readerData.line["url"][:index]
        queries = readerData.line["url"][index+1:]
    }
    if c.Method == "GET" {
        // 解析get请求参数
        if queries != "" {
            q := strings.Split(queries, "&")
            for _, v := range q {
                param := strings.Split(v, "=")
                c.queries[param[0]] = param[1]
            }
        }
    } else {
        // 判断content-type类型是不是 application/json
        contentTypes, isExist := c.request.headers["CONTENT-TYPE"]
        if isExist {
            cTypeArr := strings.Split(contentTypes, ";")
            if strings.EqualFold(cTypeArr[0], "application/json") {
                // 解析post请求参数
                json.Unmarshal([]byte(readerData.body), &(c.posts))
            }
        }
    }
}

// 获取get请求参数
func (c *Context) Query(name string) string {
    val, isExist := c.queries[name]
    if isExist {
        return val
    }
    return ""
}

// 获取post请求参数
func (c *Context) Post(name string) string {
    val, isExist := c.posts[name]
    if isExist {
        return val
    }
    return ""
}

// 获取get请求参数
func (c *Context) DefaultQuery(name, def string) string {
    val, isExist := c.queries[name]
    if isExist {
        return val
    }
    return def
}

// 获取post请求参数
func (c *Context) DefaultPost(name, def string) string {
    val, isExist := c.posts[name]
    if isExist {
        return val
    }
    return def
}

// 获取请求头
func (c *Context) GetHeader(name string) string {
    val, isExist := c.posts[strings.ToUpper(name)]
    if isExist {
        return val
    }
    return ""
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
