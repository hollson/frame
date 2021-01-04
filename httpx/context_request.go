package httpx

import (
    "bytes"
    "fmt"
    "net"
    "strconv"
    "strings"
)

type request struct {
    Path    string
    Method  string
    headers map[string]string
    queries map[string]string
    posts   map[string]string
}

type reader struct {
    conn net.Conn
    message
    buff    []byte
    buffLen int
    start   int
    end     int
}

// 实例化
func newReader(conn net.Conn, buffLen int) *reader {
    return &reader{
        conn: conn,
        message: message{
            line:   make(map[string]string),
            header: make(map[string]string),
        },
        buffLen: buffLen,
        buff:    make([]byte, buffLen),
    }
}

// 读取并解析请求行
func (reader *reader) parseLine() (isOK bool, err error) {
    index := bytes.Index(reader.buff, []byte{byte('\r'), byte('\n')})
    if index == -1 {
        // 没有解析到\r\n返回继续读取
        return
    }
    // 读取请求行
    requestLine := string(reader.buff[:index])
    arr := strings.Split(requestLine, " ")
    if len(arr) != 3 {
        return false, fmt.Errorf("bad request line")
    }
    reader.line["method"] = arr[0]
    reader.line["url"] = arr[1]
    reader.line["version"] = arr[2]

    reader.start = index + 2
    return true, nil
}

// 读取并解析请求头
func (reader *reader) parseHeader() {
    if reader.start == reader.end {
        return
    }
    index := bytes.Index(reader.buff[reader.start:], []byte{byte('\r'), byte('\n'), byte('\r'), byte('\n')})
    if index == -1 {
        return
    }
    headerStr := string(reader.buff[reader.start : reader.start+index])
    requestHeader := strings.Split(headerStr, "\r\n")
    for _, v := range requestHeader {
        arr := strings.Split(v, ":")
        if len(arr) < 2 {
            continue
        }
        reader.header[strings.ToUpper(arr[0])] = strings.ToLower(strings.Trim(strings.Join(arr[1:], ":"), " "))
    }
    reader.start += index + 4
}

// 读取并解析请求体
func (reader *reader) parseBody() (isOk bool, err error) {
    // 判断请求头中是否指明了请求体的数据长度
    contentLenStr, ok := reader.header["CONTENT-LENGTH"]
    if !ok {
        return false, fmt.Errorf("bad request:no content-length")
    }
    contentLen, err := strconv.ParseInt(contentLenStr, 10, 64)
    if err != nil {
        return false, fmt.Errorf("parse content-length error:%s", contentLenStr)
    }
    if contentLen > int64(reader.end-reader.start) {
        // 请求体长度不够，返回继续读取
        return false, nil
    }
    reader.body = string(reader.buff[reader.start : int64(reader.start)+contentLen])
    return true, nil
}

// 前移上一次未处理完的数据
func (reader *reader) move() {
    if reader.start == 0 {
        return
    }
    copy(reader.buff, reader.buff[reader.start:reader.end])
    reader.end -= reader.start
    reader.start = 0
}

// 读取http请求
func (reader *reader) responseHandler(accept chan message) (err error) {
    for {
        if reader.end == reader.buffLen {
            // 缓冲区的容量存不了一条请求的数据
            return fmt.Errorf("request is too large:%v", reader)
        }

        buffLen, err := reader.conn.Read(reader.buff)
        if err != nil {
            // 连接关闭了
            return err
        }
        reader.end += buffLen

        // 解析请求行
        isOk, err := reader.parseLine()
        if err == nil && !isOk {
            continue
        }

        if err != nil {
            return err
        }

        // 解析请求头
        reader.parseHeader()
        // 如果是post请求，解析请求体
        if len(reader.header) > 0 && strings.EqualFold(strings.ToUpper(reader.line["method"]), "POST") {
            isOk, err := reader.parseBody()
            if err != nil {
                return err
            }
            // 读取http请求体未成功
            if !isOk {
                reader.start = 0
                reader.line = make(map[string]string)
                reader.header = make(map[string]string)
                continue
            }
        }
        accept <- reader.message
        reader.move()
    }
}
