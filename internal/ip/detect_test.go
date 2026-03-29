package ip

import (
	"net"
	"testing"
)

func TestIPValidation(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1.2.3.4", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"203.0.113.42", true},
		{"2001:db8::1", true},
		{"not-an-ip", false},
		{"999.999.999.999", false},
		{"", false},
		{"1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := net.ParseIP(tt.input) != nil
			if got != tt.valid {
				t.Errorf("net.ParseIP(%q): got valid=%v, want valid=%v", tt.input, got, tt.valid)
			}
		})
	}
}
