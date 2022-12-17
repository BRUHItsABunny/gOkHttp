package gokhttp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/client"
	"github.com/BRUHItsABunny/gOkHttp/multipart"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/BRUHItsABunny/gOkHttp/responses"
	"io"
	"net/http"
	"net/url"
)

const httpBin = "https://httpbin.org/"

// This function makes a basic `*http.Client` object
func ExampleNewHTTPClient() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}
	// Just to prevent unused error in IDE
	fmt.Println(hClient)
}

// This function makes a basic `*http.Client` object that uses a proxy
func ExampleNewHTTPClientWithProxy() {
	hClient, err := NewHTTPClient(client.NewProxyOption("http://127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	// Just to prevent unused error in IDE
	fmt.Println(hClient)
}

// This function makes a basic `*http.Client` object that ensures certificate pinning
func ExampleNewHTTPClientWithPinning() {
	pinner := client.NewSSLPinningOption()
	err := pinner.AddPin("github.com", true, "sha256\\/3ftdeWqIAONye/CeEQuLGvtlw4MPnQmKgyPLugFbK8=")
	if err != nil {
		panic(err)
	}

	hClient, err := NewHTTPClient(pinner)
	if err != nil {
		panic(err)
	}
	// Just to prevent unused error in IDE
	fmt.Println(hClient)
}

// This function executes a basic GET request without any special features
func ExampleMakeGETRequest() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// GET request, no extra headers, no extra parameters
	req, err := requests.MakeGETRequest(context.Background(), httpBin+"get")
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {},
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Host": "httpbin.org",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d000f-41e34d8c1e39d9536cf7cad8"
	//  },
	//  "origin": "",
	//  "url": "https://httpbin.org/get"
	//}
}

// This function executes a basic GET request with headers and parameters
func ExampleMakeGETRequestAdvanced() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// GET request
	req, err := requests.MakeGETRequest(context.Background(), httpBin+"get",
		requests.NewHeaderOption(http.Header{"header1": {"bunnyHeader"}}),
		requests.NewURLParamOption(url.Values{"param1": {"bunnyParam"}}),
	)
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {
	//    "param1": "bunnyParam"
	//  },
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Header1": "bunnyHeader",
	//    "Host": "httpbin.org",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d003a-3f25f6d516fe495e1b51b3a9"
	//  },
	//  "origin": "",
	//  "url": "https://httpbin.org/get?param1=bunnyParam"
	//}
}

// This function executes a basic GET request with headers and parameters with a slice of options
func ExampleMakeGETRequestAdvancedSlice() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	opts := []requests.Option{
		requests.NewHeaderOption(http.Header{"header1": {"bunnyHeader"}}),
		requests.NewURLParamOption(url.Values{"param1": {"bunnyParam"}}),
	}
	// GET request
	req, err := requests.MakeGETRequest(context.Background(), httpBin+"get", opts...)
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {
	//    "param1": "bunnyParam"
	//  },
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Header1": "bunnyHeader",
	//    "Host": "httpbin.org",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d003a-3f25f6d516fe495e1b51b3a9"
	//  },
	//  "origin": "",
	//  "url": "https://httpbin.org/get?param1=bunnyParam"
	//}
}

// This function executes a basic POST request without any special features
func ExampleMakePOSTRequest() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// POST request, no extra headers, no extra parameters
	req, err := requests.MakePOSTRequest(context.Background(), httpBin+"post")
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {},
	//  "data": "",
	//  "files": {},
	//  "form": {},
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Content-Length": "0",
	//    "Host": "httpbin.org",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d00e1-5ae92df60a7950711adf3834"
	//  },
	//  "json": null,
	//  "origin": "",
	//  "url": "https://httpbin.org/post"
	//}
}

// This function executes a form POST request with headers
func ExampleMakePOSTRequestForm() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// POST request
	req, err := requests.MakePOSTRequest(context.Background(), httpBin+"post",
		requests.NewHeaderOption(http.Header{"header1": {"bunnyHeader"}}),
		requests.NewPOSTFormOption(url.Values{"param1": {"bunnyParam"}}),
	)
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {},
	//  "data": "",
	//  "files": {},
	//  "form": {
	//    "param1": "bunnyParam"
	//  },
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Content-Type": "application/x-www-form-urlencoded",
	//    "Header1": "bunnyHeader",
	//    "Host": "httpbin.org",
	//    "Transfer-Encoding": "chunked",
	//    "User-Agent": "Go-http-client/2.0",
	//
	//    "X-Amzn-Trace-Id": "Root=1-639d019a-55d2e19733b6692b398cd849"
	//  },
	//  "json": null,
	//  "origin": "",
	//  "url": "https://httpbin.org/post"
	//}
}

// This function executes a JSON POST request with headers
func ExampleMakePOSTRequestJSON() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// POST request
	req, err := requests.MakePOSTRequest(context.Background(), httpBin+"post",
		requests.NewHeaderOption(http.Header{"header1": {"bunnyHeader"}}),
		requests.NewPOSTRawOption(bytes.NewBufferString("{\"param1\":\"bunny1\"}"), "application/json"),
	)
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {},
	//  "data": "{\"param1\":\"bunny1\"}",
	//  "files": {},
	//  "form": {},
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Content-Type": "application/json",
	//    "Header1": "bunnyHeader",
	//    "Host": "httpbin.org",
	//    "Transfer-Encoding": "chunked",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d0241-2d82f2ec3ff9447b7f44e79f"
	//  },
	//  "json": {
	//    "param1": "bunny1"
	//  },
	//  "origin": "",
	//  "url": "https://httpbin.org/post"
	//}
}

// This function executes a multipart POST request with headers
func ExampleMakePOSTRequestMultipart() {
	hClient, err := NewHTTPClient()
	if err != nil {
		panic(err)
	}

	// Set up the multipart form
	multiWrapper := multipart.NewMultiPartWrapper()
	if err != nil {
		panic(err)
	}
	part, err := multiWrapper.CreateFormFile("file1", "test.txt")
	if err != nil {
		panic(err)
	}

	fakeFile := bytes.NewBufferString("content1")
	_, err = io.Copy(part, fakeFile)
	if err != nil {
		panic(err)
	}
	err = multiWrapper.Close()
	if err != nil {
		panic(err)
	}

	// POST request
	req, err := requests.MakePOSTRequest(context.Background(), httpBin+"post",
		requests.NewHeaderOption(http.Header{"header1": {"bunnyHeader"}}),
		requests.NewPOSTMultipartOption(multiWrapper),
	)
	if err != nil {
		panic(err)
	}

	// Execute it as usual
	resp, err := hClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response data, this also automatically closes the body
	respBytes, err := responses.ResponseBytes(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(respBytes))
	// Output: {
	//  "args": {},
	//  "data": "",
	//  "files": {
	//    "file1": "content1"
	//  },
	//  "form": {},
	//  "headers": {
	//    "Accept-Encoding": "gzip",
	//    "Content-Type": "multipart/form-data; boundary=093df3a49089cd04640131ebc943bf8cf1b098869fe27bbc394e5c66cbdb",
	//    "Header1": "bunnyHeader",
	//    "Host": "httpbin.org",
	//    "Transfer-Encoding": "chunked",
	//    "User-Agent": "Go-http-client/2.0",
	//    "X-Amzn-Trace-Id": "Root=1-639d0488-1aac11d157eaef7463fd4086"
	//  },
	//  "json": null,
	//  "origin": "",
	//  "url": "https://httpbin.org/post"
	//}
}
