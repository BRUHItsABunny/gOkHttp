package multipart

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
)

type Wrapper struct {
	writer *bytes.Buffer
	multi  *multipart.Writer
}

func NewMultiPartWrapper() *Wrapper {
	result := &Wrapper{}

	result.writer = &bytes.Buffer{}
	result.multi = multipart.NewWriter(result.writer)

	return result
}

func (w *Wrapper) Writer() *bytes.Buffer {
	return w.writer
}

func (w *Wrapper) FormDataContentType() string {
	return w.multi.FormDataContentType()
}

func (w *Wrapper) Boundary() string {
	return w.multi.Boundary()
}

func (w *Wrapper) Close() error {
	return w.multi.Close()
}

func (w *Wrapper) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	return w.multi.CreatePart(header)
}

func (w *Wrapper) CreateFormField(fieldName string) (io.Writer, error) {
	return w.multi.CreateFormField(fieldName)
}

func (w *Wrapper) CreateFormFile(fieldName, fileName string) (io.Writer, error) {
	return w.multi.CreateFormFile(fieldName, fileName)
}

func (w *Wrapper) SetBoundary(boundary string) error {
	return w.multi.SetBoundary(boundary)
}

func (w *Wrapper) WriteField(fieldName, value string) error {
	return w.multi.WriteField(fieldName, value)
}
