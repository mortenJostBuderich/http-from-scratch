package http1

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	var startline []string
	header := make(map[string]string)
	var body io.Reader

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("Connection closed by remote host")
			break
		} else if err != nil {
			fmt.Println("Error reading line:", err)
			return nil, err
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		if len(line) == 0 {
			if isBodyAllowed(startline[0]) {
				if clStr, ok := header["Content-Length"]; ok {
					if cl, err := strconv.Atoi(clStr); err == nil {
						body = io.LimitReader(reader, int64(cl))
					}
				}
			}
			break
		}

		if len(startline) == 0 {
			startline = strings.Fields(line)
		} else {
			before, after, found := strings.Cut(line, ":")
			if found {
				header[before] = strings.TrimSpace(after)
			}
		}
	}

	req, err := http.NewRequest(startline[0], startline[1], body)
	if err != nil {
		return nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	return req, err
}

func isBodyAllowed(startLine string) bool {
	startLine = strings.ToUpper(startLine)
	if strings.HasPrefix(startLine, "POST") ||
		strings.HasPrefix(startLine, "PUT") ||
		strings.HasPrefix(startLine, "PATCH") {
		return true
	}
	return false
}
