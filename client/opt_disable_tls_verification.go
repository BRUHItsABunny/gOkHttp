package gokhttp_client

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/http2"
)

type DisableTLSVerification struct {
}

func (o *DisableTLSVerification) Execute(client *http.Client) error {
	_, ok := client.Transport.(*http.Transport)
	if ok {
		client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	_, ok = client.Transport.(*http2.Transport)
	if ok {
		client.Transport.(*http2.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}
