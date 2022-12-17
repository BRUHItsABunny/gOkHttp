package responses

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
)

func CheckHTTPCode(resp *http.Response, expectedCodes ...int) error {
	for _, expectedCode := range expectedCodes {
		if resp.StatusCode == expectedCode {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("unexpected http status code: %d", resp.StatusCode))
}

func ResponseBytes(resp *http.Response) ([]byte, error) {
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Close: %w", err)
	}
	return respBytes, nil
}

func ResponseJSON(resp *http.Response, result interface{}) error {
	respBytes, err := ResponseBytes(resp)
	if err != nil {
		return fmt.Errorf("ResponseBytes: %w", err)
	}

	err = json.Unmarshal(respBytes, result)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}
	return nil
}

func ResponseHTML(resp *http.Response) (*goquery.Document, error) {
	respBytes, err := ResponseBytes(resp)
	if err != nil {
		return nil, fmt.Errorf("ResponseBytes: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(respBytes))
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}
	return doc, err
}
