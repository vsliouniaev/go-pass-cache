package util

import (
	"testing"
	"time"
)

func TestHumanizeDuration(t *testing.T) {
	var cases = []struct {
		in  time.Duration
		out string
	}{
		{
			in:  time.Hour * 48,
			out: "2 days",
		},
		{
			in:  (time.Hour * 48) + time.Minute*13 + time.Second*72,
			out: "2 days 14 minutes 12 seconds",
		},
	}
	for _, tt := range cases {
		t.Run(tt.out, func(t *testing.T) {
			if actual := HumanizeDuration(tt.in); tt.out != actual {
				t.Errorf("Expected %s but got %s", tt.out, actual)
			}
		})
	}
}
