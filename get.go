package gokhttp

import (
	"net/http"
	"net/url"
)

func (c *HttpClient) MakeGETRequest(URL string, parameters url.Values, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "GET", URL, nil)
	} else {
		req, err = http.NewRequest("GET", URL, nil)
	}
	req = c.readyRequest(req)
	if checkError(err) {
		query := req.URL.Query()
		for k, v := range parameters {
			for _, e := range v {
				query.Add(k, e)
			}
		}
		req.URL.RawQuery = query.Encode()
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		return req, nil
	}
	return nil, err
}
