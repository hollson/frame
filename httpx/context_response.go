package httpx

import (
    "fmt"
    "net"
    "strconv"
)

type response struct {
    status  int
    body    []byte
    bodyLen int
    headers map[string]string
    buff    []byte
    conn    net.Conn
}

func newResponse(conn net.Conn) *response {
    return &response{
        conn:    conn,
        headers: make(map[string]string),
    }
}

// 响应行
func (r *response) writeLine() {
    line := fmt.Sprintf("HTTP/1.1 %d OK\r\n", r.status)
    r.buff = append(r.buff, []byte(line)...)
}

// 响应头
func (r *response) writeHeader() {
    r.headers["server"] = "^_^"
    r.headers["content-type"] = "application/json"
    r.headers["content-length"] = strconv.FormatInt(int64(r.bodyLen), 10)
    for k, v := range r.headers {
        r.buff = append(r.buff, []byte(fmt.Sprintf("%s: %v\r\n", k, v))...)
    }
    r.buff = append(r.buff, []byte("\r\n")...)
}

func (r *response) write() {
    r.writeLine()
    r.writeHeader()
    r.buff = append(r.buff, r.body...)
    r.conn.Write(r.buff)
}

func (r *response) writeErr(status int) {
    r.status = status
    r.body = []byte("Request Error")
    r.bodyLen = len(r.body)
    r.write()
}
