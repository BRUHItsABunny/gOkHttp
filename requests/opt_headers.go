package requests

import "net/http"

type HeaderOption struct {
	Headers http.Header
	Replace bool
}

func NewHeaderOption(headers http.Header) *HeaderOption {
	return &HeaderOption{Headers: headers, Replace: true}
}

func NewHeaderOptionFromMap(headers map[string]string) *HeaderOption {
	result := &HeaderOption{Headers: http.Header{}, Replace: true}
	for key, val := range headers {
		result.Headers.Set(key, val)
	}
	return result
}

func (o *HeaderOption) Execute(req *http.Request) error {
	if o.Replace {
		for key, val := range o.Headers {
			req.Header[key] = val
		}
	} else {
		for key, val := range o.Headers {
			for _, elem := range val {
				req.Header.Add(key, elem)
			}
		}
	}
	return nil
}
