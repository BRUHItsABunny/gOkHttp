package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/client"
	"net/http"
	"os"
)

func NewHTTPClient(options ...client.Option) (*http.Client, error) {
	return client.NewHTTPClient(options...)
}

// TestHTTPClient is a function i end up using a lot in test files to control how my http.Client is initialized depending on the ENV variables USE_PROXY and PROXY_URL
func TestHTTPClient() (*http.Client, error) {
	hClient := http.DefaultClient
	opts := []client.Option{}
	if os.Getenv("USE_PROXY") == "true" {
		opts = append(opts, client.NewProxyOption(os.Getenv("PROXY_URL")))
	}

	hClient, err := client.NewHTTPClient(opts...)
	return hClient, err
}
