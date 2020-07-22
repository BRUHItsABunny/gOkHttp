package gokhttp

import "net/http"

func (c *HttpClient) MakeGETRequest(URL string, parameters, headers map[string]string) (*http.Request, error) {
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
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		return req, nil
	}
	return nil, err
}
