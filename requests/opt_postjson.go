package gokhttp_requests

import (
	"bytes"
	"github.com/BRUHItsABunny/gOkHttp/constants"
	"io"
)

func NewPOSTJSONOption(data []byte, isUTF8 bool) *POSTRawOption {
	mime := gokhttp_constants.MIMEApplicationJSON
	if isUTF8 {
		mime += "; charset=UTF-8"
	}
	return NewPOSTRawOption(bytes.NewBuffer(data), mime, int64(len(data)))
}

func NewPOSTJSONOptionFromReader(data io.Reader, length int, isUTF8 bool) *POSTRawOption {
	mime := gokhttp_constants.MIMEApplicationJSON
	if isUTF8 {
		mime += "; charset=UTF-8"
	}
	dataRc, ok := data.(io.ReadCloser)
	if !ok {
		dataRc = io.NopCloser(data)
	}
	return NewPOSTRawOption(dataRc, mime, int64(length))
}
