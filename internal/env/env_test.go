package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseFile_BasicKeyValue(t *testing.T) {
	path := writeTempEnv(t, "FOO=bar\nBAZ=qux\n")
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Key != "FOO" || entries[0].Value != "bar" {
		t.Errorf("got %+v, want FOO=bar", entries[0])
	}
	if entries[1].Key != "BAZ" || entries[1].Value != "qux" {
		t.Errorf("got %+v, want BAZ=qux", entries[1])
	}
}

func TestParseFile_DoubleQuoted(t *testing.T) {
	path := writeTempEnv(t, `KEY="hello world"`)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Value != "hello world" {
		t.Errorf("got %q, want %q", entries[0].Value, "hello world")
	}
}

func TestParseFile_SingleQuoted(t *testing.T) {
	path := writeTempEnv(t, `KEY='hello world'`)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Value != "hello world" {
		t.Errorf("got %q, want %q", entries[0].Value, "hello world")
	}
}

func TestParseFile_ExportPrefix(t *testing.T) {
	path := writeTempEnv(t, "export API_KEY=secret123\n")
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Key != "API_KEY" || entries[0].Value != "secret123" {
		t.Errorf("got %+v, want API_KEY=secret123", entries[0])
	}
}

func TestParseFile_CommentsAndEmptyLines(t *testing.T) {
	content := `# Comment line
FOO=bar

# Another comment

BAZ=qux
`
	path := writeTempEnv(t, content)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}

func TestParseFile_InvalidFormat(t *testing.T) {
	path := writeTempEnv(t, "INVALID_LINE\n")
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestParseFile_InvalidFormatDoesNotEchoLine(t *testing.T) {
	// A line without "=" may be a bare secret; it must not appear in the error.
	path := writeTempEnv(t, "sk-secret-value\n")
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if strings.Contains(err.Error(), "sk-secret-value") {
		t.Errorf("error leaks line content: %v", err)
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error should include the line number: %v", err)
	}
}

func TestParseFile_ValueWithEquals(t *testing.T) {
	path := writeTempEnv(t, "URL=https://example.com?foo=bar&baz=qux\n")
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Value != "https://example.com?foo=bar&baz=qux" {
		t.Errorf("got %q", entries[0].Value)
	}
}

func TestParseFile_EmptyFile(t *testing.T) {
	path := writeTempEnv(t, "")
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("got %d entries, want 0", len(entries))
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParseFile_InlineComments(t *testing.T) {
	content := `QUOTED="cred://work/TOKEN" # jira
SINGLE='value' # note
UNQUOTED=cred://work/TOKEN # jira
TAB=value	# tab before hash
HASH_IN_VALUE=abc#def
HASH_IN_QUOTES="a # b"
ESCAPED="a\"b"
`
	path := writeTempEnv(t, content)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"QUOTED":         "cred://work/TOKEN",
		"SINGLE":         "value",
		"UNQUOTED":       "cred://work/TOKEN",
		"TAB":            "value",
		"HASH_IN_VALUE":  "abc#def",
		"HASH_IN_QUOTES": "a # b",
		"ESCAPED":        `a\"b`,
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d", len(entries), len(want))
	}
	for _, e := range entries {
		if e.Value != want[e.Key] {
			t.Errorf("%s: got %q, want %q", e.Key, e.Value, want[e.Key])
		}
	}
}

func TestParseFile_InvalidKey(t *testing.T) {
	for _, line := range []string{"=value", "MY KEY=1", "1ABC=x", "KEY-NAME=x"} {
		path := writeTempEnv(t, line+"\n")
		_, err := ParseFile(path)
		if err == nil {
			t.Errorf("%q: expected error for invalid key", line)
			continue
		}
		if !strings.Contains(err.Error(), "line 1") {
			t.Errorf("%q: error should include the line number: %v", line, err)
		}
	}
}
