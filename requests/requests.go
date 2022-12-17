package requests

import (
	"context"
	"fmt"
	"net/http"
)

// MakeGETRequest Makes a GET request
func MakeGETRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	// Force context
	if ctx == nil {
		ctx = context.Background()
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakeHEADRequest Makes a HEAD request
func MakeHEADRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	// Force context
	if ctx == nil {
		ctx = context.Background()
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, httpURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakePOSTRequest Makes a POST request
func MakePOSTRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	// Force context
	if ctx == nil {
		ctx = context.Background()
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, httpURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}
