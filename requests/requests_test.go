package gokhttp_requests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/multipart"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
)

const (
	testURL = "http://127.0.0.1/"
	httpBin = "https://httpbin.org/"
)

func debugDump(reqRaw []byte) {
	fmt.Println("===")
	fmt.Print(string(reqRaw))
	fmt.Println("===")
}

func TestRequests(t *testing.T) {

	params := url.Values{
		"param1": {"bunny1"},
	}
	headers := http.Header{
		"header1": {"bunnyHead1"},
	}
	fakeFile := bytes.NewBufferString("content1")
	fakeJSON := "{\"param1\":\"bunny1\"}"

	t.Run("MakeHEAD", func(t *testing.T) {
		expected := []byte("HEAD / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")

		req, err := MakeHEADRequest(context.Background(), testURL)
		require.NoError(t, err, "requests.MakeHEADRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, false)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakeHEADRequest: raw request not as expected")
	})

	t.Run("MakeGET", func(t *testing.T) {
		expected := []byte("GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")

		req, err := MakeGETRequest(context.Background(), testURL)
		require.NoError(t, err, "requests.MakeGETRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakeGETRequest: raw request not as expected")
	})

	t.Run("MakeGET_Advanced", func(t *testing.T) {
		expected := []byte("GET /?param1=bunny1 HTTP/1.1\r\nHost: 127.0.0.1\r\nheader1: bunnyHead1\r\n\r\n")

		req, err := MakeGETRequest(context.Background(), testURL, NewURLParamOption(params), NewHeaderOption(headers))
		require.NoError(t, err, "requests.MakeGETRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakeGETRequest: raw request not as expected")
	})

	t.Run("MakePOST", func(t *testing.T) {
		expected := []byte("POST / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")

		req, err := MakePOSTRequest(context.Background(), testURL)
		require.NoError(t, err, "requests.MakePOSTRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakePOSTRequest: raw request not as expected")
	})

	t.Run("MakePOST_Form", func(t *testing.T) {
		expected := []byte("POST / HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nparam1=bunny1")

		req, err := MakePOSTRequest(context.Background(), testURL, NewPOSTFormOption(params))
		require.NoError(t, err, "requests.MakePOSTRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakePOSTRequest: raw request not as expected")
	})

	t.Run("MakePOST_Multipart", func(t *testing.T) {
		expected := []byte("POST / HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Type: multipart/form-data; boundary=561b690102f1b14e82f9acb028b32bbce087aeda3d343c1236d25db02377\r\n\r\n--561b690102f1b14e82f9acb028b32bbce087aeda3d343c1236d25db02377\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\nContent-Type: application/octet-stream\r\n\r\ncontent1\r\n--561b690102f1b14e82f9acb028b32bbce087aeda3d343c1236d25db02377--\r\n")

		multiWrapper := gokhttp_multipart.NewMultiPartWrapper()
		// Set boundary is required since it gets randomized when empty
		err := multiWrapper.SetBoundary("561b690102f1b14e82f9acb028b32bbce087aeda3d343c1236d25db02377")
		require.NoError(t, err, "multiWrapper.SetBoundary: errored unexpectedly.")
		part, err := multiWrapper.CreateFormFile("file", "test.txt")
		require.NoError(t, err, "multiWrapper.CreateFormFile: errored unexpectedly.")
		_, err = io.Copy(part, fakeFile)
		require.NoError(t, err, "io.Copy: errored unexpectedly.")
		err = multiWrapper.Close()
		require.NoError(t, err, "multiWrapper.Close: errored unexpectedly.")

		req, err := MakePOSTRequest(context.Background(), testURL, NewPOSTMultipartOption(multiWrapper))
		require.NoError(t, err, "requests.MakePOSTRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakePOSTRequest: raw request not as expected")
	})

	t.Run("MakePOST_Raw", func(t *testing.T) {
		expected := []byte("POST / HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Type: application/json\r\n\r\n{\"param1\":\"bunny1\"}")

		req, err := MakePOSTRequest(context.Background(), testURL, NewPOSTRawOption(bytes.NewBufferString(fakeJSON), "application/json", int64(len(fakeJSON))))
		require.NoError(t, err, "requests.MakePOSTRequest: errored unexpectedly.")
		reqRaw, err := httputil.DumpRequest(req, true)
		require.NoError(t, err, "httputil.DumpRequest: errored unexpectedly.")
		require.Equal(t, expected, reqRaw, "requests.MakePOSTRequest: raw request not as expected")
	})
}
