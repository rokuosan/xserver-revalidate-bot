package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateRecord(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("Accept", "text/html")
	req.RemoteAddr = "127.0.0.1:12345"

	record := createRecord(req)

	if record.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent to be 'TestAgent/1.0', got '%s'", record.UserAgent)
	}

	if record.Method != "GET" {
		t.Errorf("Expected Method to be 'GET', got '%s'", record.Method)
	}

	if record.RequestURI != "/test?param=value" {
		t.Errorf("Expected RequestURI to be '/test?param=value', got '%s'", record.RequestURI)
	}

	if record.RemoteAddr != "127.0.0.1:12345" {
		t.Errorf("Expected RemoteAddr to be '127.0.0.1:12345', got '%s'", record.RemoteAddr)
	}

	if record.Headers["User-Agent"] != "TestAgent/1.0" {
		t.Errorf("Expected User-Agent header to be 'TestAgent/1.0', got '%s'", record.Headers["User-Agent"])
	}

	if record.Headers["Accept"] != "text/html" {
		t.Errorf("Expected Accept header to be 'text/html', got '%s'", record.Headers["Accept"])
	}
}

func TestFormatHeaders(t *testing.T) {
	headers := map[string]string{
		"User-Agent": "TestAgent/1.0",
		"Accept":     "text/html",
		"Host":       "localhost:8080",
	}

	formatted := formatHeaders(headers)

	if !strings.Contains(formatted, "User-Agent: TestAgent/1.0") {
		t.Error("Expected formatted headers to contain User-Agent")
	}

	if !strings.Contains(formatted, "Accept: text/html") {
		t.Error("Expected formatted headers to contain Accept")
	}

	if !strings.Contains(formatted, "Host: localhost:8080") {
		t.Error("Expected formatted headers to contain Host")
	}
}

func TestServer_HandleUA(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest("GET", "/api/ua", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")

	rr := httptest.NewRecorder()

	server.HandleUA(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestAgent/1.0") {
		t.Error("Expected response to contain TestAgent/1.0")
	}
}

func TestServer_HandleHeaders(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()

	server.HandleHeaders(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestAgent/1.0") {
		t.Error("Expected response to contain TestAgent/1.0")
	}

	if !strings.Contains(body, "application/json") {
		t.Error("Expected response to contain Accept header value")
	}
}

func TestUserAgentRecord_TimestampIsRecent(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	record := createRecord(req)

	now := time.Now()
	diff := now.Sub(record.Timestamp)

	if diff > time.Second {
		t.Errorf("Expected timestamp to be recent, but it was %v ago", diff)
	}
}
