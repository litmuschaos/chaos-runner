package utils

import (
	"os"
	"testing"
	"time"
)

func TestDurationFromEnv(t *testing.T) {
	const key = "TEST_CONTROL_PLANE_HEARTBEAT_DURATION"
	tests := []struct {
		name, value    string
		fallback, want time.Duration
	}{
		{name: "unset uses fallback", value: "", fallback: 30 * time.Second, want: 30 * time.Second},
		{name: "Go duration", value: "45s", fallback: time.Second, want: 45 * time.Second},
		{name: "integer seconds", value: "12", fallback: time.Second, want: 12 * time.Second},
		{name: "invalid uses fallback", value: "invalid", fallback: 7 * time.Second, want: 7 * time.Second},
		{name: "nonpositive uses fallback", value: "0s", fallback: 9 * time.Second, want: 9 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, tt.value)
			}
			if got := durationFromEnv(key, tt.fallback); got != tt.want {
				t.Fatalf("durationFromEnv() = %s, want %s", got, tt.want)
			}
		})
	}
	_ = os.Unsetenv(key)
}
