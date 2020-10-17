package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"time"
)

var EmptyMap map[string]string
var DefaultHTTPTimeout = time.Second * time.Duration(30)

// proxyUrl, _ := url.Parse("http://127.0.0.1:8866")
var DefaultGOKHTTPOptions = &HttpClientOptions{
	JarOptions: &cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: false, Filename: ".cookies", EncryptionPassword: ""},
	Transport: &http.Transport{
		TLSHandshakeTimeout: DefaultHTTPTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		// Proxy:http.ProxyURL(proxyUrl),
	},
	RefererOptions: &RefererOptions{Update: true, Use: true},
	Timeout:        &DefaultHTTPTimeout,
}
