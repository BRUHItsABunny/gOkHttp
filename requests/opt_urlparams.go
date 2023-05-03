package gokhttp_requests

import (
	"net/http"
	"net/url"
)

type URLParamOption struct {
	Values url.Values
}

func NewURLParamOption(values url.Values) *URLParamOption {
	return &URLParamOption{Values: values}
}

func NewURLParamOptionFromMap(values map[string]string) *URLParamOption {
	finalValues := url.Values{}
	for key, val := range values {
		finalValues.Add(key, val)
	}
	return &URLParamOption{Values: finalValues}
}

func (o *URLParamOption) Execute(req *http.Request) error {
	query := req.URL.Query()
	for key, val := range o.Values {
		for _, elem := range val {
			query.Add(key, elem)
		}
	}
	req.URL.RawQuery = query.Encode()
	return nil
}
