// Package env parses .env files for cred import and cred run --env-file.
package env

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Entry represents a single key-value pair from a .env file.
type Entry struct {
	Key   string
	Value string
}

// ParseFile reads a .env file and returns its entries.
// Supports:
//   - KEY=VALUE
//   - KEY="VALUE" (double-quoted, strips quotes)
//   - KEY='VALUE' (single-quoted, strips quotes)
//   - KEY=VALUE # comment (a "#" after whitespace or after the closing quote starts a comment)
//   - # comments
//   - empty lines
//   - export KEY=VALUE (strips export prefix)
func ParseFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip "export " prefix
		line = strings.TrimPrefix(line, "export ")

		// Split on first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid format (expected KEY=VALUE)", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		if !validKey.MatchString(key) {
			return nil, fmt.Errorf("line %d: invalid key (expected letters, digits and '_', not starting with a digit)", lineNum)
		}

		entries = append(entries, Entry{Key: key, Value: parseValue(strings.TrimSpace(parts[1]))})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	return entries, nil
}

// validKey matches names that are valid environment variable names.
var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// parseValue strips surrounding quotes and a trailing comment ("#" after
// whitespace, or after the closing quote) from a trimmed raw value.
func parseValue(raw string) string {
	if len(raw) >= 2 && (raw[0] == '"' || raw[0] == '\'') {
		q := raw[0]
		// Prefer the last closing quote so values with escaped quotes keep
		// their content; a later quote inside the comment falls back to an earlier one.
		for end := len(raw) - 1; end > 0; end-- {
			if raw[end] != q {
				continue
			}
			rest := strings.TrimSpace(raw[end+1:])
			if rest == "" || strings.HasPrefix(rest, "#") {
				return raw[1:end]
			}
		}
		return raw
	}

	for i := 1; i < len(raw); i++ {
		if raw[i] == '#' && (raw[i-1] == ' ' || raw[i-1] == '\t') {
			return strings.TrimSpace(raw[:i])
		}
	}
	return raw
}
