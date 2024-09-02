package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length uint
	}{
		{"default", 5},
		{"short", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomString(tt.length)
			require.Equal(t, int(tt.length), len(got))
		})
	}
}

func TestBuildData(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "empty", args: "", want: "N/A"},
		{"default", "v0.0.1", "v0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildData(tt.args); got != tt.want {
				t.Errorf("BuildData() = %v, want %v", got, tt.want)
			}
		})
	}
}
