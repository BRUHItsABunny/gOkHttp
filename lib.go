package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/client"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

func NewHTTPClient(options ...client.Option) (*http.Client, error) {
	return client.NewHTTPClient(options...)
}

// TestHTTPClient is a function i end up using a lot in test files to control how my http.Client is initialized depending on the .env file in my project
func TestHTTPClient() (*http.Client, error) {
	hClient := http.DefaultClient
	err := godotenv.Load(".env")
	if err != nil {
		return hClient, err
	}

	opts := []client.Option{}
	if os.Getenv("USE_PROXY") == "true" {
		opts = append(opts, client.NewProxyOption(os.Getenv("PROXY_URL")))
	}

	hClient, err = client.NewHTTPClient(opts...)
	return hClient, err
}
