package http1

import (
	"fmt"
	"net"
	"net/http"
)

type Response struct {
	conn       net.Conn
	headers    http.Header
	statusCode int
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		conn:    conn,
		headers: make(http.Header),
	}
}

func (r *Response) Header() http.Header {
	return r.headers
}

func (r *Response) Write(data []byte) (n int, err error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	statusText := http.StatusText(r.statusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.statusCode, statusText)
	r.conn.Write([]byte(statusLine))
	for key, values := range r.headers {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			r.conn.Write([]byte(headerLine))
		}
	}
	r.conn.Write([]byte("\r\n"))
	return r.conn.Write(data)
}

func (r *Response) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
