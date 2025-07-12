package xserver

import (
	"testing"
)

func Test_mustJoinURL(t *testing.T) {
	tests := []struct {
		base   string
		elems  []string
		expect string
	}{
		{
			base:   "https://example.com",
			elems:  []string{"path", "to", "resource"},
			expect: "https://example.com/path/to/resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			result := mustJoinURL(tt.base, tt.elems...)
			if result.String() != tt.expect {
				t.Errorf("expected %s, got %s", tt.expect, result.String())
			}
		})
	}
}
