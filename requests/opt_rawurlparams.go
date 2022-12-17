package requests

import (
	"net/http"
)

type RawURLParamOption struct {
	RawQuery string
}

func NewRawURLParamOption(rawQuery string) *RawURLParamOption {
	return &RawURLParamOption{RawQuery: rawQuery}
}

func (o *RawURLParamOption) Execute(req *http.Request) {
	req.URL.RawQuery = o.RawQuery
}
