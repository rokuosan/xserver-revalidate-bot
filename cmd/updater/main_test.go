package main

import (
	"strings"
	"testing"
)

func Test_parseHeaderFile(t *testing.T) {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
	}

	reader := strings.NewReader(`{"User-Agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"}`)
	parsedHeaders, err := parseHeaderFile(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for key, value := range headers {
		if parsedHeaders[key] != value {
			t.Errorf("Expected %s for key %s, got %s", value, key, parsedHeaders[key])
		}
	}
}

func Test_getHeaders(t *testing.T) {
	headers, err := getHeaders()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(headers) == 0 {
		t.Error("Expected headers to be non-empty")
	}

	if _, exists := headers["User-Agent"]; !exists {
		t.Error("Expected User-Agent header to be present")
	}
}
