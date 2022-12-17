package client

import (
	"golang.org/x/net/http2"
	"net/http"
	"time"
)

type TimeOutOption struct {
	ClientTimeout         time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	ExpectContinueTimeout time.Duration
	WriteByteTimeout      time.Duration
	ReadIdleTimeout       time.Duration
	PingTimeout           time.Duration
}

func NewTimeOutOption(client, handshake, responseHeader, idleConn, expectContinue, writeByte, readIdle, ping time.Duration) *TimeOutOption {
	return &TimeOutOption{
		ClientTimeout:         client,
		TLSHandshakeTimeout:   handshake,
		ResponseHeaderTimeout: responseHeader,
		IdleConnTimeout:       idleConn,
		ExpectContinueTimeout: expectContinue,
		WriteByteTimeout:      writeByte,
		ReadIdleTimeout:       readIdle,
		PingTimeout:           ping,
	}
}

func (o *TimeOutOption) Execute(client *http.Client) error {
	client.Timeout = o.ClientTimeout

	_, ok := client.Transport.(*http.Transport)
	if ok {
		client.Transport.(*http.Transport).TLSHandshakeTimeout = o.TLSHandshakeTimeout
		client.Transport.(*http.Transport).ResponseHeaderTimeout = o.ResponseHeaderTimeout
		client.Transport.(*http.Transport).IdleConnTimeout = o.IdleConnTimeout
		client.Transport.(*http.Transport).ExpectContinueTimeout = o.ExpectContinueTimeout
	}

	_, ok = client.Transport.(*http2.Transport)
	if ok {
		client.Transport.(*http2.Transport).WriteByteTimeout = o.WriteByteTimeout
		client.Transport.(*http2.Transport).ReadIdleTimeout = o.ReadIdleTimeout
		client.Transport.(*http2.Transport).PingTimeout = o.PingTimeout
	}
	return nil
}
