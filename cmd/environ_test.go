package cmd

import (
	"testing"
)

func TestSetEnv(t *testing.T) {
	tests := []struct {
		name     string
		environ  []string
		key      string
		value    string
		expected []string
	}{
		{
			name:     "overwrite existing variable",
			environ:  []string{"HOME=/home/user", "API_KEY=old", "PATH=/usr/bin"},
			key:      "API_KEY",
			value:    "new",
			expected: []string{"HOME=/home/user", "API_KEY=new", "PATH=/usr/bin"},
		},
		{
			name:     "append when not present",
			environ:  []string{"HOME=/home/user", "PATH=/usr/bin"},
			key:      "API_KEY",
			value:    "secret",
			expected: []string{"HOME=/home/user", "PATH=/usr/bin", "API_KEY=secret"},
		},
		{
			name:     "prefix safety: API_KEY does not match API_KEY_V2",
			environ:  []string{"API_KEY_V2=other"},
			key:      "API_KEY",
			value:    "secret",
			expected: []string{"API_KEY_V2=other", "API_KEY=secret"},
		},
		{
			name:     "empty value",
			environ:  []string{"API_KEY=old"},
			key:      "API_KEY",
			value:    "",
			expected: []string{"API_KEY="},
		},
		{
			name:     "empty environ",
			environ:  []string{},
			key:      "API_KEY",
			value:    "secret",
			expected: []string{"API_KEY=secret"},
		},
		{
			name:     "overwrite first entry",
			environ:  []string{"API_KEY=old", "OTHER=val"},
			key:      "API_KEY",
			value:    "new",
			expected: []string{"API_KEY=new", "OTHER=val"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy input to avoid mutating test data across runs
			input := make([]string, len(tt.environ))
			copy(input, tt.environ)

			result := setEnv(input, tt.key, tt.value)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d\n  got:  %v\n  want: %v", len(result), len(tt.expected), result, tt.expected)
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("index %d: got %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
