package gokhttp

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *HttpClient) MakePOSTRequest(URL string, postFields, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	form := url.Values{}
	for k, v := range postFields {
		form.Add(k, v)
	}
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "POST", URL, strings.NewReader(form.Encode()))
	} else {
		req, err = http.NewRequest("POST", URL, strings.NewReader(form.Encode()))
	}
	req = c.readyRequest(req)
	if checkError(err) {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakePOSTRequestWithParameters(URL string, parameters url.Values, postFields, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	form := url.Values{}
	for k, v := range postFields {
		form.Add(k, v)
	}
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "POST", URL, strings.NewReader(form.Encode()))
	} else {
		req, err = http.NewRequest("POST", URL, strings.NewReader(form.Encode()))
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
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakeMultiPartPOSTRequest(URL, contentType string, postBody io.Reader, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "POST", URL, postBody)
	} else {
		req, err = http.NewRequest("POST", URL, postBody)
	}
	req = c.readyRequest(req)
	if checkError(err) {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		req.Header.Add("Content-Type", contentType)
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakeMultiPartPOSTRequestWithParameters(URL string, parameters url.Values, contentType string, postBody io.Reader, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "POST", URL, postBody)
	} else {
		req, err = http.NewRequest("POST", URL, postBody)
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
		req.Header.Add("Content-Type", contentType)
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakeRawPOSTRequest(URL string, parameters url.Values, postBody io.Reader, headers map[string]string) (*http.Request, error) {
	/*
		No content type -> post body will be seen as DATA
		x-wwww-form content type -> post body will be seen as FIELDS
		multipart content type -> post body will be seen as files + fields or fields
	*/
	var (
		req *http.Request
		err error
	)
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "POST", URL, postBody)
	} else {
		req, err = http.NewRequest("POST", URL, postBody)
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
