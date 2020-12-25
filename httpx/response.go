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
func (response *response) writeLine() {
    line := fmt.Sprintf("HTTP/1.1 %d OK\r\n", response.status)
    response.buff = append(response.buff, []byte(line)...)
}

// 响应头
func (response *response) writeHeader() {
    response.headers["server"] = "^_^"
    response.headers["content-type"] = "application/json"
    response.headers["content-length"] = strconv.FormatInt(int64(response.bodyLen), 10)
    for k, v := range response.headers {
        response.buff = append(response.buff, []byte(fmt.Sprintf("%s: %v\r\n", k, v))...)
    }
    response.buff = append(response.buff, []byte("\r\n")...)
}

func (response *response) write() {
    response.writeLine()
    response.writeHeader()
    response.buff = append(response.buff, response.body...)
    response.conn.Write(response.buff)
}

func (response *response) errWrite(status int) {
    response.status = status
    response.body = []byte("Request Error")
    response.bodyLen = len(response.body)
    response.write()
}
