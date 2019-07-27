package gokhttp

import (
	"context"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"net"
	"net/http"
	"time"
)

type HttpClient struct {
	Client         *http.Client
	RefererOptions RefererOptions
	Headers        map[string]string
	Context        *context.Context
}

type RefererOptions struct {
	Update bool
	Use    bool
	Value  string
}

type HttpClientOptions struct {
	JarOptions        *cookies.JarOptions
	Transport         *http.Transport
	Timeout           *time.Duration
	SSLPinningOptions *SSLPinner
	RefererOptions    *RefererOptions
	RedirectPolicy    func(req *http.Request, via []*http.Request) error
	Headers           map[string]string
	Context           *context.Context
}

type HttpResponse struct {
	*http.Response
}

type HttpJSONResponse struct {
	data []byte
}

type Dialer func(network, addr string) (net.Conn, error)

type SSLPin struct {
	SkipCA    bool
	Pins      []string //sha256
	Hostname  string
	Algorithm string
}

type SSLPinner struct {
	SSLPins map[string]SSLPin
}
