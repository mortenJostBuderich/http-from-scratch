package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			err := HandleConnection(c, handler)
			if err != nil {
				fmt.Printf("error handling connection: %v\n", err)
			}
		}(con)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		req, err := ParseRequest(reader)

		if err != nil {
			if err == io.EOF {
				return nil
			} else if errors.Is(err, BadVersion) {
				res := NewResponse(conn)
				res.Header().Set("Connection", "close")
				res.WriteHeader(http.StatusHTTPVersionNotSupported)
				_, _ = res.Write([]byte("505 \r\n"))
				return nil
			}
			return err
		}
		mux.ServeHTTP(NewResponse(conn), req)
		if req.Close {
			break
		}
	}
	return nil
}
