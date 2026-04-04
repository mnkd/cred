package cmd

import "strings"

// setEnv sets key=value in the environ slice, removing all existing
// entries with the same key and appending the new value at the end.
// This ensures exactly one entry per key, matching envchain's setenv(overwrite=1).
func setEnv(environ []string, key, value string) []string {
	prefix := key + "="
	out := environ[:0]
	for _, entry := range environ {
		if !strings.HasPrefix(entry, prefix) {
			out = append(out, entry)
		}
	}
	return append(out, prefix+value)
}
