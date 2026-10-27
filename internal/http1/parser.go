package http1

import (
	"bufio"
	"errors"
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
			if clStr, ok := header["Content-Length"]; ok {
				if cl, err := strconv.Atoi(clStr); err == nil {
					body = io.LimitReader(reader, int64(cl))
				}
			}
			break
		} else if len(startline) == 0 {
			err := verifyStartingLine(line)
			if err != nil {
				return nil, err
			}
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

func verifyStartingLine(startingLine string) error {
	if !strings.HasSuffix(startingLine, "\n\r") ||
		len(strings.Fields(startingLine)) != 3 ||
		strings.Count(startingLine, " ") != 2 {
		return errors.New("invalid starting line")
	}
	return nil
}
