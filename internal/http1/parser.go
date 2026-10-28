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
	method, path, majorVersion, minorVersion, err := parseStartLine(reader)
	if err != nil {
		return nil, err
	}

	header, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}

	body, err := parseBody(reader, header)
	if err != nil {
		return nil, err
	}

	// create request
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Method = method
	req.RequestURI = path
	req.ProtoMajor = majorVersion
	req.ProtoMinor = minorVersion
	req.Header = header

	connHeader := strings.ToLower(req.Header.Get("Connection"))
	if req.ProtoMajor == 1 && req.ProtoMinor == 0 {
		req.Close = connHeader != "keep-alive"
	} else {
		req.Close = connHeader == "close"
	}

	return req, nil
}

func parseStartLine(reader *bufio.Reader) (string, string, int, int, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", "", 0, 0, err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}

	line = strings.TrimSuffix(line, "\r\n")
	startlineFields := strings.Fields(line)
	if len(startlineFields) != 3 {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}

	method := startlineFields[0]
	if strings.ToUpper(method) != "GET" &&
		method != "POST" &&
		method != "PUT" &&
		method != "DELETE" &&
		method != "OPTIONS" &&
		method != "PATCH" &&
		method != "HEAD" {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}

	path := startlineFields[1]

	versionString := startlineFields[2]
	before, after, found := strings.Cut(versionString, "/")
	if !found || before != "HTTP" {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}
	majorStr, minorStr, hasDot := strings.Cut(after, ".")
	if !hasDot {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}
	major, err1 := strconv.Atoi(majorStr)
	minor, err2 := strconv.Atoi(minorStr)
	if err1 != nil || err2 != nil {
		return "", "", 0, 0, fmt.Errorf("malformed")
	}

	return method, path, major, minor, err
}

func parseHeaders(reader *bufio.Reader) (http.Header, error) {
	header := make(http.Header)
	for {
		line, prefix, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}
		if prefix {
			return nil, fmt.Errorf("header line is too long: %d", line)
		}
		if string(line) == "" {
			break
		}
		if strings.HasPrefix(string(line), " ") {
			return nil, fmt.Errorf("header lines may not start with a space: %s", line)
		}
		if strings.Contains(string(line), "\r") {
			return nil, fmt.Errorf("header lines may not contain CR: %s", line)
		}
		parts := strings.SplitN(string(line), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header line: %s", line)
		}
		if len(parts[0]) == 0 {
			return nil, fmt.Errorf("invalid header line: %s", line)
		}
		if strings.Contains(parts[0], " ") {
			return nil, fmt.Errorf("invalid header line: %s", line)
		}
		splits := strings.Split(parts[1], ",")
		for i := range splits {
			header.Add(strings.TrimSpace(parts[0]), strings.ToLower(strings.TrimSpace(splits[i])))
		}
	}
	return header, nil
}

func parseBody(reader *bufio.Reader, header http.Header) (io.Reader, error) {
	var body io.Reader
	if cl, err := strconv.Atoi(header.Get("Content-Length")); err == nil {
		body = io.LimitReader(reader, int64(cl))
	}
	return body, nil
}
