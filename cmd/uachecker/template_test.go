package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplateRenderer_Render(t *testing.T) {
	renderer := NewTemplateRenderer()

	err := renderer.LoadTemplate("templates/index.html")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	data := TemplateData{
		UserAgent:        "TestAgent/1.0",
		RemoteAddr:       "127.0.0.1:12345",
		Method:           "GET",
		RequestURI:       "/test",
		Timestamp:        "2023-01-01 12:00:00",
		HeadersFormatted: "User-Agent: TestAgent/1.0\nAccept: text/html\n",
	}

	var buf bytes.Buffer
	err = renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "TestAgent/1.0") {
		t.Error("Expected output to contain TestAgent/1.0")
	}

	if !strings.Contains(output, "127.0.0.1:12345") {
		t.Error("Expected output to contain remote address")
	}

	if !strings.Contains(output, "User-Agent Checker") {
		t.Error("Expected output to contain page title")
	}
}

func TestTemplateRenderer_RenderWithoutTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()

	data := TemplateData{
		UserAgent: "TestAgent/1.0",
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)

	if err == nil {
		t.Error("Expected error when rendering without loaded template")
	}
}

func TestTemplateRenderer_LoadNonExistentTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()

	err := renderer.LoadTemplate("nonexistent.html")

	if err == nil {
		t.Error("Expected error when loading non-existent template")
	}
}
