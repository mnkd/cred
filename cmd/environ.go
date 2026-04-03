package cmd

import "strings"

// setEnv sets key=value in the environ slice, overwriting any existing
// entry with the same key. If the key does not exist, it is appended.
// This matches the behavior of envchain's setenv(overwrite=1).
func setEnv(environ []string, key, value string) []string {
	prefix := key + "="
	for i, entry := range environ {
		if strings.HasPrefix(entry, prefix) {
			environ[i] = prefix + value
			return environ
		}
	}
	return append(environ, prefix+value)
}
