package llmkit

import (
	"os"
	"testing"
)

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  *float64
	}{
		{"valid float", "TEST_FLOAT", "0.7", ptr(0.7)},
		{"zero", "TEST_FLOAT", "0", ptr(0.0)},
		{"negative", "TEST_FLOAT", "-0.5", ptr(-0.5)},
		{"empty", "TEST_FLOAT", "", nil},
		{"invalid", "TEST_FLOAT", "not-a-number", nil},
		{"unset", "TEST_FLOAT_UNSET", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := parseFloat(tt.key)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseFloat() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("parseFloat() = nil, want %v", *tt.want)
				} else if *got != *tt.want {
					t.Errorf("parseFloat() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  *int
	}{
		{"valid int", "TEST_INT", "4096", intPtr(4096)},
		{"zero", "TEST_INT", "0", intPtr(0)},
		{"negative", "TEST_INT", "-1", intPtr(-1)},
		{"empty", "TEST_INT", "", nil},
		{"invalid", "TEST_INT", "not-a-number", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := parseInt(tt.key)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseInt() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("parseInt() = nil, want %v", *tt.want)
				} else if *got != *tt.want {
					t.Errorf("parseInt() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  *int64
	}{
		{"valid int64", "TEST_INT64", "12345", int64Ptr(12345)},
		{"large value", "TEST_INT64", "9223372036854775807", int64Ptr(9223372036854775807)},
		{"empty", "TEST_INT64", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := parseInt64(tt.key)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseInt64() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("parseInt64() = nil, want %v", *tt.want)
				} else if *got != *tt.want {
					t.Errorf("parseInt64() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{"valid string", "TEST_STRING", "medium", "medium"},
		{"empty", "TEST_STRING", "", ""},
		{"unset", "TEST_STRING_UNSET", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := parseString(tt.key)
			if got != tt.want {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for pointers
func ptr(v float64) *float64 { return &v }
func intPtr(v int) *int      { return &v }
func int64Ptr(v int64) *int64 { return &v }
