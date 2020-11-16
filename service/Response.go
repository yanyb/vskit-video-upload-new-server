package service

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	Protocol                  = "HTTP/1.1"
	CRLF                      = "\r\n"
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusRequestTimeout      = http.StatusRequestTimeout
	StatusInternalServerError = http.StatusInternalServerError
	StatusBadGateway          = http.StatusBadGateway
)

type Response struct {
	Code    int
	Headers map[string]string
	Body    string
}

func (res *Response) String() string {
	content := []string{}
	content = append(content, fmt.Sprintf("%s %d", Protocol, res.Code))
	contentLength := len(res.Body)
	res.Headers["Content-Length"] = fmt.Sprintf("%d", contentLength)
	for k, v := range res.Headers {
		content = append(content, k+": "+v)
	}
	// 空行
	content = append(content, "")
	content = append(content, res.Body)
	return strings.Join(content, CRLF)
}
