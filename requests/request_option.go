package gokhttp_requests

import "net/http"

type Option interface {
	Execute(r *http.Request) error
}

func ExecuteOpts(req *http.Request, opts ...Option) error {
	for _, opt := range opts {
		err := opt.Execute(req)
		if err != nil {
			return err
		}
	}
	return nil
}
