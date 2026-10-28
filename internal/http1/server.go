package http1

import (
	"bufio"
	"fmt"
	"httpServer/internal/mux1"
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
		handler = mux1.NewMyMux()
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
		res := NewResponse(conn)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.SendHeaders()
			break
		}
		if req == nil {
			break
		}

		if !(req.ProtoMajor == 1 && req.ProtoMinor == 1) && !(req.ProtoMajor == 1 && req.ProtoMinor == 0) {
			res.WriteHeader(http.StatusHTTPVersionNotSupported)
			res.SendHeaders()
			break
		}

		mux.ServeHTTP(res, req)
		if err = res.SendHeaders(); err != nil {
			return err
		}

		if req.Close {
			break
		}
	}
	return nil
}
