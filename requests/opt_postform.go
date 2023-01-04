package requests

import (
	"net/url"
	"strings"
)

func NewPOSTFormOption(values url.Values) *POSTRawOption {
	return NewPOSTRawOption(strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", int64(len(values.Encode())))
}

func NewPOSTFormOptionFromMap(values map[string]string) *POSTRawOption {
	finalValues := url.Values{}
	for key, val := range values {
		finalValues.Add(key, val)
	}
	return NewPOSTFormOption(finalValues)
}
