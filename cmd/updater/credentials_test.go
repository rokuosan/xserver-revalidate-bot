package main

import "testing"

func Test_maskCredentials(t *testing.T) {
	tests := []struct {
		credential string
		expected   string
	}{
		{"1234", "****"},
		{"123456", "12****56"},
	}
	for _, tt := range tests {
		t.Run(tt.credential, func(t *testing.T) {
			got := maskCredential(tt.credential)
			if got != tt.expected {
				t.Errorf("maskCredential(%q) = %q; want %q", tt.credential, got, tt.expected)
			}
		})
	}
}
