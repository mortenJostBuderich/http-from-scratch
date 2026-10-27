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
		trimmedLine := strings.TrimSuffix(line, "\r\n")

		if len(startline) == 0 {
			err := verifyStartingLine(line)
			if err != nil {
				return nil, err
			}
			startline = strings.Fields(trimmedLine)
		} else if len(trimmedLine) == 0 {
			if clStr, ok := header["Content-Length"]; ok {
				if cl, err := strconv.Atoi(clStr); err == nil {
					body = io.LimitReader(reader, int64(cl))
				}
			}
			break
		} else {
			before, after, found := strings.Cut(trimmedLine, ":")
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
	if startline[2] == "HTTP/1.0" {
		req.Proto = "HTTP/1.0"
		req.ProtoMajor = 1
		req.ProtoMinor = 0
	}
	connHeader := strings.ToLower(req.Header.Get("Connection"))
	if req.ProtoMinor == 0 {
		req.Close = connHeader != "keep-alive"
	} else {
		req.Close = connHeader == "close"
	}
	return req, nil
}

func verifyStartingLine(startline string) error {
	if !strings.HasSuffix(startline, "\r\n") {
		return MalFormed
	}
	startline = strings.TrimSuffix(startline, "\r\n")

	startlineFields := strings.Fields(startline)
	if len(startlineFields) != 3 {
		return MalFormed
	}

	method := startlineFields[0]
	if strings.ToUpper(method) != "GET" &&
		method != "POST" &&
		method != "PUT" &&
		method != "DELETE" &&
		method != "OPTIONS" &&
		method != "PATCH" &&
		method != "HEAD" {
		return MalFormed
	}

	versionString := startlineFields[2]
	before, after, found := strings.Cut(versionString, "/")

	// Check the prefix
	if !found || before != "HTTP" {
		return MalFormed
	}
	majorStr, minorStr, hasDot := strings.Cut(after, ".")
	if !hasDot {
		return MalFormed
	}
	major, err1 := strconv.Atoi(majorStr)
	minor, err2 := strconv.Atoi(minorStr)
	if err1 != nil || err2 != nil {
		return MalFormed
	}
	if major != 1 {
		return BadVersion
	}
	if minor != 0 && minor != 1 {
		return BadVersion
	}
	return nil
}

var (
	BadVersion = errors.New("bad version")
	MalFormed  = errors.New("malformed")
)
