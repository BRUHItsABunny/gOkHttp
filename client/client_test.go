package gokhttp_client

import (
	"context"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TODO: Add cookiejar option
func TestNewSSLPinningOption(t *testing.T) {
	pinner := NewSSLPinningOption()
	err := pinner.AddPin("github.com", true, "sha256\\/3ftdeWqIAONye/CeEQuLGvtlw4MPnQmKgyPLugFbK8=")
	require.NoError(t, err, "pinner.AddPin: errored unexpectedly.")

	hClient, err := NewHTTPClient(pinner)
	require.NoError(t, err, "NewHTTPClient: errored unexpectedly.")

	req, err := gokhttp_requests.MakeGETRequest(context.Background(), "https://github.com")
	require.NoError(t, err, "requests.MakeGETRequest: errored unexpectedly.")

	_, err = hClient.Do(req)
	require.NoError(t, err, "hClient.Do: errored unexpectedly.")
}

func TestNewJarOption(t *testing.T) {
	jar, err := gokhttp_cookies.NewCookieJar(".cookies", "test", nil)
	require.NoError(t, err, "cookies.NewCookieJar: errored unexpectedly.")

	err = jar.Load()
	require.NoError(t, err, "jar.Load: errored unexpectedly.")

	hClient, err := NewHTTPClient(NewJarOption(jar), NewProxyOption("http://127.0.0.1:8888"))
	require.NoError(t, err, "NewHTTPClient: errored unexpectedly.")

	for i := 0; i <= 1; i++ {
		req, err := gokhttp_requests.MakeGETRequest(context.Background(), "https://google.com")
		require.NoError(t, err, "requests.MakeGETRequest: errored unexpectedly.")

		_, err = hClient.Do(req)
		require.NoError(t, err, "hClient.Do: errored unexpectedly.")

		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	err = jar.Save()
	require.NoError(t, err, "jar.Save: errored unexpectedly.")
}
