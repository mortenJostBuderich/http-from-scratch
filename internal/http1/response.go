package http1

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

type Response struct {
	conn        net.Conn
	headers     http.Header
	statusCode  int
	wroteHeader bool
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

func (r *Response) WriteHeader(statusCode int) {
	if r.wroteHeader || r.statusCode != 0 {
		return
	}
	r.statusCode = statusCode
}

func (r *Response) Write(data []byte) (n int, err error) {
	if !r.wroteHeader {
		if r.statusCode == 0 {
			r.statusCode = http.StatusOK
		}
		if r.headers.Get("Content-Length") == "" {
			r.headers.Set("Content-Length", strconv.Itoa(len(data)))
		}
		if err := r.sendHeaders(); err != nil {
			return 0, err
		}
	}
	return r.conn.Write(data)
}

func (r *Response) sendHeaders() error {
	r.wroteHeader = true
	statusText := http.StatusText(r.statusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.statusCode, statusText)
	if _, err := r.conn.Write([]byte(statusLine)); err != nil {
		return err
	}
	for key, values := range r.headers {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			if _, err := r.conn.Write([]byte(headerLine)); err != nil {
				return err
			}
		}
	}
	_, err := r.conn.Write([]byte("\r\n"))
	return err
}

func (r *Response) SendHeaders() error {
	if !r.wroteHeader {
		if r.statusCode == 0 {
			r.statusCode = http.StatusOK
		}
		if r.headers.Get("Content-Length") == "" {
			r.headers.Set("Content-Length", "0")
		}
		return r.sendHeaders()
	}
	return nil
}
