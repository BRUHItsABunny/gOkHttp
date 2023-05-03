package gokhttp_requests

import gokhttp_multipart "github.com/BRUHItsABunny/gOkHttp/multipart"

func NewPOSTMultipartOption(writer *gokhttp_multipart.Wrapper) *POSTRawOption {
	return NewPOSTRawOption(writer.Writer(), writer.FormDataContentType(), int64(writer.Writer().Len()))
}
