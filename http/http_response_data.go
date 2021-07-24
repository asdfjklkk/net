package http

import (
	"net/http"
)

type HttpResponseData struct {
	Error         error
	Status        string
	StatusCode    int
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        http.Header
	ContentLength int64
	Body          []byte
	Request       *http.Request
}
