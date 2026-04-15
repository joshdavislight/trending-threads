package main

import (
	"testing"
	"time"
)

func TestPostTriggerDebounceDuration(t *testing.T) {
	tests := []struct {
		name string
		secs int
		want time.Duration
	}{
		{"zero defaults to two seconds", 0, 2 * time.Second},
		{"negative defaults to two seconds", -3, 2 * time.Second},
		{"one second", 1, 1 * time.Second},
		{"ten seconds", 10, 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := postTriggerDebounceDuration(tt.secs); got != tt.want {
				t.Fatalf("postTriggerDebounceDuration(%d) = %v, want %v", tt.secs, got, tt.want)
			}
		})
	}
}
