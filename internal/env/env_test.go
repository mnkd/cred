package env

import (
	"os"
	"path/filepath"
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
	content := `# コメント行
FOO=bar

# もうひとつコメント

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
