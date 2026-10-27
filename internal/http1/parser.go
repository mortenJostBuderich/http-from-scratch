package http1

import (
	"bufio"
	"errors"
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
		if err != nil {
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
			startline = strings.Fields(line)
			err := verifyStartingLine(startline)
			if err != nil {
				return nil, err
			}
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
	if strings.ToUpper(req.Header.Get("Connection")) == "CLOSE" || startline[2] == "HTTP/1.0" {
		req.Close = true
	}

	return req, nil
}

func verifyStartingLine(startingLine []string) error {

	// cut

	if startingLine[2] != "HTTP/1.0" && startingLine[2] != "HTTP/1.1" {
		return BadVersion
	} else if len(startingLine) != 3 {
		return MalFormed
	}
	return nil
}

var (
	BadVersion = errors.New("bad version")
	MalFormed  = errors.New("malformed")
)
