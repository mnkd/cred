package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "default version when not set by ldflags",
			version: "dev",
			want:    "cred version dev\n",
		},
		{
			name:    "version set by ldflags",
			version: "0.1.0",
			want:    "cred version 0.1.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := version
			version = tt.version
			t.Cleanup(func() { version = original })

			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs([]string{"version"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
