package client

import "net/http"

func NewHTTPClient(options ...Option) (*http.Client, error) {
	client := &http.Client{Transport: &http.Transport{}}

	for _, option := range options {
		err := option.Execute(client)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
