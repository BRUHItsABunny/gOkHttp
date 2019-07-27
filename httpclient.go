package gokhttp

import (
	"bytes"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"golang.org/x/net/publicsuffix"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func checkError(err error) bool {
	if err == nil {
		return true
	}
	return false
}

func GetHTTPClient(o *HttpClientOptions) HttpClient {
	httpClient := HttpClient{
		Client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 15 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		Headers: map[string]string{},
	}
	cookieJar, _ := cookies.New(&cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: true})
	refOps := RefererOptions{Update: false, Use: false}
	if o != nil {
		if o.Timeout != nil {
			httpClient.Client.Timeout = *o.Timeout
		}
		if o.Transport != nil {
			httpClient.Client.Transport = o.Transport
			if o.SSLPinningOptions != nil {
				o.Transport.DialTLS = MakeDialer(*o.SSLPinningOptions)
			}
		}
		if o.JarOptions != nil {
			cookieJar, _ = cookies.New(o.JarOptions)
		}
		if o.RefererOptions != nil {
			refOps = *o.RefererOptions
		}
		if o.Headers != nil {
			httpClient.Headers = o.Headers
		}
		if o.RedirectPolicy != nil {
			httpClient.Client.CheckRedirect = o.RedirectPolicy
		}
	}
	httpClient.Client.Jar = cookieJar
	httpClient.RefererOptions = refOps
	return httpClient
}

func (c *HttpClient) readyRequest(req *http.Request) *http.Request {
	if c.RefererOptions.Use && c.RefererOptions.Value != "" {
		req.Header.Add("Referer", c.RefererOptions.Value)
	}
	if c.RefererOptions.Update {
		c.RefererOptions.Value = req.URL.String()
	}
	for k, v := range c.Headers {
		req.Header.Add(k, v)
	}
	return req
}

func (c *HttpClient) MakeGETRequest(URL string, parameters, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
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

func (c *HttpClient) MakePOSTRequest(URL string, postFields, headers map[string]string) (*http.Request, error) {
	form := url.Values{}
	for k, v := range postFields {
		form.Add(k, v)
	}
	req, err := http.NewRequest("POST", URL, strings.NewReader(form.Encode()))
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

func (c *HttpClient) MakePOSTRequestWithParameters(URL string, postFields, parameters, headers map[string]string) (*http.Request, error) {
	form := url.Values{}
	for k, v := range postFields {
		form.Add(k, v)
	}
	req, err := http.NewRequest("POST", URL, strings.NewReader(form.Encode()))
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
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakeMultiPartPOSTRequest(URL, contentType string, postBody io.Reader, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest("POST", URL, postBody)
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

func (c *HttpClient) MakeMultiPartPOSTRequestWithParameters(URL, contentType string, postBody io.Reader, parameters, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest("POST", URL, postBody)
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
		req.Header.Add("Content-Type", contentType)
		return req, nil
	}
	return nil, err
}

func (c *HttpClient) MakeRawPOSTRequest(URL string, postBody io.Reader, parameters, headers map[string]string) (*http.Request, error) {
	/*
		No content type -> post body will be seen as DATA
		x-wwww-form content type -> post body will be seen as FIELDS
		multipart content type -> post body will be seen as files + fields or fields
	*/
	req, err := http.NewRequest("POST", URL, postBody)
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

func (c *HttpClient) MakePostFields(postFields map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range postFields {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}

func (c *HttpClient) MakePostMultiPart(postFields map[string]string, files map[string]io.Reader, fieldNames []string) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	i := 0
	for k, v := range files {
		part, err := writer.CreateFormFile(fieldNames[i], k)
		if !checkError(err) {
			return nil, "", err
		}
		_, err = io.Copy(part, v)
		if !checkError(err) {
			return nil, "", err
		}
		i++
	}
	for key, val := range postFields {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if checkError(err) {
		return body, writer.FormDataContentType(), nil
	}
	return nil, "", err
}

func (c *HttpClient) Do(req *http.Request) (*HttpResponse, error) {
	response, err := c.Client.Do(req)
	if checkError(err) {
		if c.Client.Jar.(*cookies.Jar).Persistent() {
			err = c.Client.Jar.(*cookies.Jar).Save()
			if checkError(err) {
				return &HttpResponse{response}, err
			}
			return nil, err
		}
		return &HttpResponse{response}, err
	}
	return nil, err
}
