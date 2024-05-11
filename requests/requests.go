package gokhttp_requests

import (
	"context"
	"fmt"
	"net/http"
)

// MakeGETRequest Makes a GET request
func MakeGETRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodGet, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
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
	req, err := baseRequest(ctx, http.MethodHead, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
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
	req, err := baseRequest(ctx, http.MethodPost, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakePUTRequest Makes a PUT request
func MakePUTRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodPut, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakePATCHRequest Makes a PATCH request
func MakePATCHRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodPatch, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakeDELETERequest Makes a DELETE request
func MakeDELETERequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodDelete, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakeCONNECTRequest Makes a CONNECT request
func MakeCONNECTRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodConnect, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakeOPTIONSRequest Makes a OPTIONS request
func MakeOPTIONSRequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodOptions, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

// MakeTRACERequest Makes a TRACE request
func MakeTRACERequest(ctx context.Context, httpURL string, opts ...Option) (*http.Request, error) {
	req, err := baseRequest(ctx, http.MethodTrace, httpURL)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	// Execute options
	err = ExecuteOpts(req, opts...)
	if err != nil {
		return nil, fmt.Errorf("ExecuteOpts: %w", err)
	}

	return req, nil
}

func baseRequest(ctx context.Context, method, httpURL string) (*http.Request, error) {
	// Force context
	if ctx == nil {
		ctx = context.Background()
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, method, httpURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	return req, nil
}
