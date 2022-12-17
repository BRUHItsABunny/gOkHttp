package client

import "net/http"

type Option interface {
	Execute(client *http.Client) error
}
