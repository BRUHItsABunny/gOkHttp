package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/client"
	"net/http"
)

func NewHTTPClient(options ...client.Option) (*http.Client, error) {
	return client.NewHTTPClient(options...)
}
