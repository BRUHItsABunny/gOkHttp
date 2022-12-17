package requests

import (
	"io"
	"net/http"
)

type POSTRawOption struct {
	Body        io.ReadCloser
	ContentType string
}

func NewPOSTRawOption(data io.Reader, contentType string) *POSTRawOption {
	dataRc, ok := data.(io.ReadCloser)
	if !ok {
		dataRc = io.NopCloser(data)
	}
	return &POSTRawOption{Body: dataRc, ContentType: contentType}
}

func (o *POSTRawOption) Execute(req *http.Request) error {
	if req.Method == http.MethodPost {
		req.Body = o.Body
		req.Header.Set("Content-Type", o.ContentType)
	}
	return nil
}
