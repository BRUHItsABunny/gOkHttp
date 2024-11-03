package gokhttp_client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	"golang.org/x/net/http2"
)

type TLSConfigOption interface {
	Execute(client *http.Client) error
	ExecuteTLSConfig(config *tls.Config) error
}

func executeTLSConfig(hClient *http.Client, tlsConfigOpt TLSConfigOption) error {
	typedH1Trans, ok := hClient.Transport.(*http.Transport)
	if ok {
		if typedH1Trans.TLSClientConfig == nil {
			typedH1Trans.TLSClientConfig = &tls.Config{}
		}
		err := tlsConfigOpt.ExecuteTLSConfig(typedH1Trans.TLSClientConfig)
		if err != nil {
			return err
		}
	}
	typedH2Trans, ok := hClient.Transport.(*http2.Transport)
	if ok {
		if typedH2Trans.TLSClientConfig == nil {
			typedH2Trans.TLSClientConfig = &tls.Config{}
		}
		err := tlsConfigOpt.ExecuteTLSConfig(typedH2Trans.TLSClientConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

type RawTLSConfigOption struct {
	Config *tls.Config
}

func (opt *RawTLSConfigOption) ExecuteV1(client *http.Client) error {
	typedH1Trans, ok := client.Transport.(*http.Transport)
	if ok {
		typedH1Trans.TLSClientConfig = opt.Config
	}
	typedH2Trans, ok := client.Transport.(*http2.Transport)
	if ok {
		typedH2Trans.TLSClientConfig = opt.Config
	}
	return nil
}

func (opt *RawTLSConfigOption) Execute(client *http.Client) error {
	return opt.processTransport(client, 1)
}

func (opt *RawTLSConfigOption) ExecuteTLSConfig(config *tls.Config) error {
	config = opt.Config
	return nil
}

func (opt *RawTLSConfigOption) processTransport(clientOrTransport any, depth int) error {
	var (
		ok                  bool
		clientOrTransportRV reflect.Value
	)
	clientOrTransportRV, ok = clientOrTransport.(reflect.Value)
	if !ok {
		clientOrTransportRV = reflect.ValueOf(clientOrTransport)
	}

	if clientOrTransportRV.Kind() != reflect.Ptr || clientOrTransportRV.IsNil() {
		return fmt.Errorf("expected non-nil pointer to struct")
	}

	// fmt.Println(fmt.Sprintf("Depth %d clientOrTransport.Type: %s", depth, clientOrTransport.Type().String()))
	clientOrTransportElem := clientOrTransportRV.Elem()
	// fmt.Println(fmt.Sprintf("Depth %d rvElem.Type: %s", depth, clientOrTransportElem.Type().String()))

	// http.Client, oohttp.Client, http.Transport, oohttp.StdlibTransport, oohttp.Transport are all structs.
	if clientOrTransportElem.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct")
	}

	// Check for ".Transport" field
	transportField := clientOrTransportElem.FieldByName("Transport")
	// Should always be an interface
	if transportField.Kind() == reflect.Interface {
		transportField = transportField.Elem()
	}

	// If it exists, we should recurse into it
	if transportField.IsValid() {
		// fmt.Println(fmt.Sprintf("Depth %d configField.Type: %s", depth, transportField.Type().String()))
		if transportField.Kind() == reflect.Ptr {
			if transportField.IsNil() {
				return fmt.Errorf("transport field is nil")
			}
			return opt.processTransport(transportField, depth+1)
		} else if transportField.Kind() == reflect.Struct {
			return opt.processTransport(transportField.Addr(), depth+1)
		} else {
			return fmt.Errorf("transport field is not a struct or pointer to struct")
		}
	} else {
		// Set ".TLSClientConfig" field
		tlsConfigField := clientOrTransportElem.FieldByName("TLSClientConfig")
		if tlsConfigField.IsValid() {
			if !tlsConfigField.CanSet() {
				return fmt.Errorf("cannot set TLSClientConfig field")
			}

			tlsConfigField.Set(reflect.ValueOf(opt.Config))

			return nil
		} else {
			return fmt.Errorf("TLSClientConfig field not found")
		}
	}
}

func NewRawTLSConfigOption(config *tls.Config) *RawTLSConfigOption {
	return &RawTLSConfigOption{Config: config}
}

type DisableTLSVerificationOption struct{}

func (opt *DisableTLSVerificationOption) Execute(client *http.Client) error {
	return executeTLSConfig(client, opt)
}

func (opt *DisableTLSVerificationOption) ExecuteTLSConfig(config *tls.Config) error {
	config.InsecureSkipVerify = true
	return nil
}

func NewDisableTLSVerificationOption() *DisableTLSVerificationOption {
	return &DisableTLSVerificationOption{}
}

type MTLSOption struct {
	CAs          *x509.CertPool
	Certificates []tls.Certificate
}

func (opt *MTLSOption) Execute(client *http.Client) error {
	return executeTLSConfig(client, opt)
}

func (opt *MTLSOption) ExecuteTLSConfig(config *tls.Config) error {
	config.Certificates = opt.Certificates
	config.RootCAs = opt.CAs
	return nil
}

func (opt *MTLSOption) AddCAFromCert(ca *x509.Certificate) error {
	opt.CAs.AddCert(ca)
	return nil
}

func (opt *MTLSOption) AddCAFromPEM(pemCerts []byte) error {
	ok := opt.CAs.AppendCertsFromPEM(pemCerts)
	if !ok {
		return errors.New("failed to add ca from pem")
	}
	return nil
}

func (opt *MTLSOption) AddCAFromFile(caPath string) error {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return errors.New("failed to read ca")
	}

	ok := opt.CAs.AppendCertsFromPEM(caCert)
	if !ok {
		return errors.New("failed to add ca from pem")
	}
	return nil
}

func (opt *MTLSOption) AddClientCertFromCert(cert tls.Certificate) error {
	opt.Certificates = append(opt.Certificates, cert)
	return nil
}

func (opt *MTLSOption) AddClientCertFromPEM(certPEMBlock, keyPEMBlock []byte) error {
	clientCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return errors.New("failed to add client certificate from pem")
	}
	opt.Certificates = append(opt.Certificates, clientCert)
	return nil
}

func (opt *MTLSOption) AddClientCertFromFile(clientCertPath, clientKeyPath string) error {
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return errors.New("failed to add client certificate from path")
	}
	opt.Certificates = append(opt.Certificates, clientCert)
	return nil
}

func NewMTLSOption(caPool *x509.CertPool, certificates []tls.Certificate) *MTLSOption {
	return &MTLSOption{CAs: caPool, Certificates: certificates}
}

type TLSKeyLoggingOption struct {
	Destination io.Writer
}

func (opt *TLSKeyLoggingOption) Execute(client *http.Client) error {
	return executeTLSConfig(client, opt)
}

func (opt *TLSKeyLoggingOption) ExecuteTLSConfig(config *tls.Config) error {
	config.KeyLogWriter = opt.Destination
	return nil
}

func NewTLSKeyLoggingOption(writer io.Writer) *TLSKeyLoggingOption {
	return &TLSKeyLoggingOption{Destination: writer}
}

func NewTLSKeyLoggingOptionToFile(path string) (*TLSKeyLoggingOption, error) {
	if path == "" {
		path = fmt.Sprintf("gokhttp_keys_%d.log", time.Now().Unix())
	}
	writer, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile: %w", err)
	}
	return &TLSKeyLoggingOption{Destination: writer}, nil
}
