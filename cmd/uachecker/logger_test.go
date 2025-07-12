package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLogger_LogRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	record := UserAgentRecord{
		UserAgent:  "TestAgent/1.0",
		RemoteAddr: "127.0.0.1:12345",
		Method:     "GET",
		RequestURI: "/test",
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	logger.LogRequest(record)

	output := buf.String()

	if !strings.Contains(output, "TestAgent/1.0") {
		t.Error("Expected log to contain user agent")
	}

	if !strings.Contains(output, "127.0.0.1:12345") {
		t.Error("Expected log to contain remote address")
	}

	if !strings.Contains(output, "GET") {
		t.Error("Expected log to contain method")
	}

	if !strings.Contains(output, "/test") {
		t.Error("Expected log to contain request URI")
	}
}

func TestLogger_LogError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.LogError("test error", "additional context")

	output := buf.String()

	if !strings.Contains(output, "ERROR") {
		t.Error("Expected log to contain ERROR level")
	}

	if !strings.Contains(output, "test error") {
		t.Error("Expected log to contain error message")
	}

	if !strings.Contains(output, "additional context") {
		t.Error("Expected log to contain additional context")
	}
}

func TestLogger_LogInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.LogInfo("test info message")

	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Error("Expected log to contain INFO level")
	}

	if !strings.Contains(output, "test info message") {
		t.Error("Expected log to contain info message")
	}
}

func TestLogger_DefaultLogger(t *testing.T) {
	logger := NewLogger(nil)

	// Should not panic with nil writer (uses default)
	logger.LogInfo("test message")
}
