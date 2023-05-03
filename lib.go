package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/client"
	"net/http"
	"os"
)

func NewHTTPClient(options ...gokhttp_client.Option) (*http.Client, error) {
	return gokhttp_client.NewHTTPClient(options...)
}

// TestHTTPClient is a function i end up using a lot in test files to control how my http.Client is initialized depending on the ENV variables USE_PROXY and PROXY_URL
func TestHTTPClient(options ...gokhttp_client.Option) (*http.Client, error) {
	hClient := http.DefaultClient
	opts := []gokhttp_client.Option{}
	if os.Getenv("USE_PROXY") == "true" {
		opts = append(opts, gokhttp_client.NewProxyOption(os.Getenv("PROXY_URL")))
	}

	opts = append(opts, options...)

	hClient, err := gokhttp_client.NewHTTPClient(opts...)
	return hClient, err
}
