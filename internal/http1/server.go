package http1

import (
	"bufio"
	"fmt"
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
			defer con.Close()
			err = HandleConnection(con, handler)
			if err != nil {
				fmt.Printf("error handling connection: %v\n", err)
			}
		}(con)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	reader := bufio.NewReader(conn)
	req, err := ParseRequest(reader)
	if err != nil {
		return err
	}
	mux.ServeHTTP(NewResponse(conn), req)
	return nil
}
