package gokhttp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/client"
	"net/http"
	"os"
)

func NewHTTPClient(options ...gokhttp_client.Option) (*http.Client, error) {
	return gokhttp_client.NewHTTPClient(options...)
}

// TestHTTPClient is a function i end up using a lot in test files to control how my http.Client is initialized depending on the ENV variables USE_PROXY and PROXY_URL
func TestHTTPClient(options ...gokhttp_client.Option) (*http.Client, error) {
	hClient := http.DefaultClient
	opts := []gokhttp_client.Option{}
	if os.Getenv("USE_PROXY") == "true" {
		opts = append(opts, gokhttp_client.NewProxyOption(os.Getenv("PROXY_URL")))
	}
	if os.Getenv("USE_MTLS") == "true" {
		opt := gokhttp_client.NewMTLSOption(x509.NewCertPool(), []tls.Certificate{})
		err := opt.AddCAFromFile(os.Getenv("CA_CERT"))
		if err != nil {
			return nil, fmt.Errorf("opt.AddCAFromFile: %w", err)
		}
		err = opt.AddClientCertFromFile(os.Getenv("CLIENT_CERT"), os.Getenv("CLIENT_KEY"))
		if err != nil {
			return nil, fmt.Errorf("opt.AddClientCertFromFile: %w", err)
		}
		opts = append(opts, opt)
	}
	if os.Getenv("USE_WIRESHARK") == "true" {
		opt, err := gokhttp_client.NewTLSKeyLoggingOptionToFile(os.Getenv("WIRESHARK_LOGFILE"))
		if err != nil {
			return nil, fmt.Errorf("gokhttp_client.NewTLSKeyLoggingOptionToFile: %w", err)
		}
		opts = append(opts, opt)
	}

	opts = append(opts, options...)

	hClient, err := gokhttp_client.NewHTTPClient(opts...)
	return hClient, err
}
