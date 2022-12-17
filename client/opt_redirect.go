package client

import "net/http"

type RedirectOption struct {
}

func (o *RedirectOption) Execute(client *http.Client) error {
	client.CheckRedirect = noRedirect
	return nil
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
