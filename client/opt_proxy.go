package client

import (
	"fmt"
	"net/http"
	"net/url"
)

type ProxyOption struct {
	ProxyURL string
}

func NewProxyOption(proxyURL string) *ProxyOption {
	return &ProxyOption{ProxyURL: proxyURL}
}

// TODO: Allow for other types of proxies outside of stdlib too?

func (o *ProxyOption) Execute(client *http.Client) error {
	puo, err := url.Parse(o.ProxyURL)
	if err != nil {
		return fmt.Errorf("ProxyOption: url.Parse: %w", err)
	}
	client.Transport.(*http.Transport).Proxy = http.ProxyURL(puo)
	return nil
}
