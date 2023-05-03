package gokhttp_requests

func NewPOSTMultipartOption(writer *multipart.Wrapper) *POSTRawOption {
	return NewPOSTRawOption(writer.Writer(), writer.FormDataContentType(), int64(writer.Writer().Len()))
}
