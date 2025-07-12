package xserver

import "testing"

func Test_VPSID_String(t *testing.T) {
	tests := []struct {
		name     string
		vpsid    VPSID
		expected string
	}{
		{
			name:     "Non-empty VPSID",
			vpsid:    VPSID("vps-12345"),
			expected: "vps-12345",
		},
		{
			name:     "Empty VPSID",
			vpsid:    VPSID(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vpsid.String(); got != tt.expected {
				t.Errorf("VPSID.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_UniqueID_String(t *testing.T) {
	tests := []struct {
		name     string
		uniqueID UniqueID
		expected string
	}{
		{
			name:     "Non-empty UniqueID",
			uniqueID: UniqueID("unique-12345"),
			expected: "unique-12345",
		},
		{
			name:     "Empty UniqueID",
			uniqueID: UniqueID(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uniqueID.String(); got != tt.expected {
				t.Errorf("UniqueID.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
