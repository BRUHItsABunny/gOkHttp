package gokhttp_client

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

type ProxyOption struct {
	ProxyURL string
	proxyObj *url.URL
}

func NewProxyOption(proxyURL string) *ProxyOption {
	return &ProxyOption{ProxyURL: proxyURL}
}

func (o *ProxyOption) ExecuteV1(client *http.Client) error {
	puo, err := url.Parse(o.ProxyURL)
	if err != nil {
		return fmt.Errorf("ProxyOption: url.Parse: %w", err)
	}

	_, ok := client.Transport.(*http.Transport)
	if ok {
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(puo)
	}

	return nil
}

func (o *ProxyOption) Execute(client *http.Client) error {
	puo, err := url.Parse(o.ProxyURL)
	if err != nil {
		return fmt.Errorf("ProxyOption: url.Parse: %w", err)
	}
	o.proxyObj = puo
	return o.processTransport(client, 1)
}

func (o *ProxyOption) processTransport(clientOrTransport any, depth int) error {
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
			return o.processTransport(transportField, depth+1)
		} else if transportField.Kind() == reflect.Struct {
			return o.processTransport(transportField.Addr(), depth+1)
		} else {
			return fmt.Errorf("transport field is not a struct or pointer to struct")
		}
	} else {
		// Set ".Proxy" field
		proxyField := clientOrTransportElem.FieldByName("Proxy")
		if proxyField.IsValid() && proxyField.Kind() == reflect.Func {
			if !proxyField.CanSet() {
				return fmt.Errorf("cannot set Proxy field")
			}
			// Literally make a new function based on it's signature
			proxyFieldType := proxyField.Type()
			newFuncValue := reflect.MakeFunc(proxyFieldType, func(args []reflect.Value) (results []reflect.Value) {
				// fmt.Println("Dynamic Proxy called with arguments:")
				for i, arg := range args {
					fmt.Printf("  Arg %d: %v\n", i, arg.Interface())
				}
				results = make([]reflect.Value, proxyFieldType.NumOut())
				results[0] = reflect.ValueOf(o.proxyObj)
				results[1] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
				return results
			})
			// Set the Proxy field to the new function
			proxyField.Set(newFuncValue)
			return nil
		} else {
			return fmt.Errorf("proxy field not found or not a function")
		}
	}
}
