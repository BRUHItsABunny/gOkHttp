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
	if o == nil {
		o = DefaultGOKHTTPOptions
	}
	httpClient := HttpClient{
		Client: &http.Client{
			Timeout: *o.Timeout,
			Transport: &http.Transport{
				TLSHandshakeTimeout: *o.Timeout,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		Headers: map[string]string{},
	}
	cookieJar, _ := cookies.New(&cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: true})
	refOps := RefererOptions{Update: false, Use: false}
	if o.Timeout != nil {
		httpClient.Client.Timeout = *o.Timeout
	}
	if o.Transport != nil {
		httpClient.Client.Transport = o.Transport
		if o.SSLPinningOptions != nil {
			o.Transport.DialTLSContext = MakeDialer(*o.SSLPinningOptions)
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
	if o.Context != nil {
		httpClient.Context = o.Context
		httpClient.CancelF = o.CancelF
	}
	httpClient.Client.Jar = cookieJar
	httpClient.RefererOptions = refOps
	httpClient.ClientOptions = o
	return httpClient
}

func GetHTTPDownloadClient(o *HttpClientOptions) HttpClient {
	if o == nil {
		o = DefaultGOKHTTPOptions
	}
	httpClient := HttpClient{
		Client: &http.Client{
			Timeout: time.Duration(0), // No time-out
			Transport: &http.Transport{
				TLSHandshakeTimeout: *o.Timeout,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		Headers: map[string]string{},
	}
	cookieJar, _ := cookies.New(&cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: true})
	refOps := RefererOptions{Update: false, Use: false}
	if o.Timeout != nil {
		httpClient.Client.Timeout = time.Duration(0)
	}
	if o.Transport != nil {
		httpClient.Client.Transport = o.Transport
		if o.SSLPinningOptions != nil {
			o.Transport.DialTLSContext = MakeDialer(*o.SSLPinningOptions)
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
	if o.Context != nil {
		httpClient.Context = o.Context
	}
	httpClient.Client.Jar = cookieJar
	httpClient.RefererOptions = refOps
	httpClient.ClientOptions = o
	return httpClient
}

func (c *HttpClient) SetProxy(proxyURLStr string) error {
	proxyUrl, err := url.Parse(proxyURLStr)
	if err == nil {
		c.Client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyUrl)
	}
	return err
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
	if c.Context != nil {
		req = req.WithContext(*c.Context)
	}
	return req
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

func (c *HttpClient) CopyClient() HttpClient {
	return GetHTTPClient(c.ClientOptions)
}
