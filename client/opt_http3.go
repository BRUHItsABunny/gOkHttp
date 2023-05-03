package gokhttp_client

import (
	"github.com/quic-go/quic-go/http3"
	"net/http"
)

type HTTP3Option struct {
	Transport *http3.RoundTripper
}

func NewHTTP3Option(trans *http3.RoundTripper) *HTTP3Option {
	return &HTTP3Option{Transport: trans}
}

func (o *HTTP3Option) Execute(client *http.Client) error {
	client.Transport = o.Transport
	return nil
}
