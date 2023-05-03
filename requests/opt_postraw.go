package gokhttp_requests

import (
	"io"
	"net/http"
)

type POSTRawOption struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
}

func NewPOSTRawOption(data io.Reader, contentType string, contentLength int64) *POSTRawOption {
	dataRc, ok := data.(io.ReadCloser)
	if !ok {
		dataRc = io.NopCloser(data)
	}
	return &POSTRawOption{Body: dataRc, ContentType: contentType, ContentLength: contentLength}
}

func (o *POSTRawOption) Execute(req *http.Request) error {
	if req.Method == http.MethodPost {
		req.Body = o.Body
		req.Header.Set("Content-Type", o.ContentType)
		req.ContentLength = o.ContentLength
	}
	return nil
}
