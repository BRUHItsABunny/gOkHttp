package requests

import (
	"github.com/BRUHItsABunny/gOkHttp/multipart"
)

func NewPOSTMultipartOption(writer *multipart.Wrapper) *POSTRawOption {
	return NewPOSTRawOption(writer.Writer(), writer.FormDataContentType())
}
