package applications

import "testing"

func TestValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"applied is valid", "applied", true},
		{"interview is valid", "interview", true},
		{"offer is valid", "offer", true},
		{"rejected is valid", "rejected", true},
		{"withdrawn is valid", "withdrawn", true},
		{"unknown is invalid", "unknown", false},
		{"empty is invalid", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isValidStatus(test.status)

			if got != test.want {
				t.Errorf(
					"isValidStatus(%q) = %v, want %v",
					test.status,
					got,
					test.want,
				)
			}
		})
	}
}
