package client

import (
	"context"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO: Add cookiejar option
func TestNewSSLPinningOption(t *testing.T) {
	pinner := NewSSLPinningOption()
	err := pinner.AddPin("github.com", true, "sha256\\/3ftdeWqIAONye/CeEQuLGvtlw4MPnQmKgyPLugFbK8=")
	require.NoError(t, err, "pinner.AddPin: errored unexpectedly.")

	hClient, err := NewHTTPClient(pinner)
	require.NoError(t, err, "NewHTTPClient: errored unexpectedly.")

	req, err := requests.MakeGETRequest(context.Background(), "https://github.com")
	require.NoError(t, err, "requests.MakeGETRequest: errored unexpectedly.")

	_, err = hClient.Do(req)
	require.NoError(t, err, "hClient.Do: errored unexpectedly.")
}
