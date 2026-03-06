package db

import (
	"fmt"
	"testing"
	"time"
)

func TestIsRetryableOpenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "sqlite busy",
			err:  fmt.Errorf("database is locked (5) (SQLITE_BUSY)"),
			want: true,
		},
		{
			name: "wrapped busy",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("ping sqlite database: database is locked")),
			want: true,
		},
		{
			name: "non retryable",
			err:  fmt.Errorf("no such table: issues"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableOpenError(tc.err); got != tc.want {
				t.Fatalf("isRetryableOpenError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestOpenBackoff(t *testing.T) {
	if got := openBackoff(0); got != 20*time.Millisecond {
		t.Fatalf("attempt 0 backoff = %s, want %s", got, 20*time.Millisecond)
	}
	if got := openBackoff(3); got != 160*time.Millisecond {
		t.Fatalf("attempt 3 backoff = %s, want %s", got, 160*time.Millisecond)
	}
	if got := openBackoff(10); got != 320*time.Millisecond {
		t.Fatalf("attempt 10 backoff = %s, want %s", got, 320*time.Millisecond)
	}
}
