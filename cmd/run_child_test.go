//go:build darwin

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mnkd/cred/internal/mask"
)

const testSecret = "s3cr3t-value"

func tempFile(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func readFile(t *testing.T, f *os.File) string {
	t.Helper()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func runSh(t *testing.T, script string, stdout, stderr *os.File) int {
	t.Helper()
	code, err := runMasked("/bin/sh", []string{"sh", "-c", script},
		[]string{"S=" + testSecret}, []string{testSecret},
		childIO{Stdout: stdout, Stderr: stderr})
	if err != nil {
		t.Fatalf("runMasked: %v", err)
	}
	return code
}

func TestRunMasked_MasksStdoutAndStderr(t *testing.T) {
	stdout, stderr := tempFile(t, "out"), tempFile(t, "err")
	code := runSh(t, `echo "out=$S"; echo "err=$S" >&2`, stdout, stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := readFile(t, stdout), "out="+mask.Replacement+"\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got, want := readFile(t, stderr), "err="+mask.Replacement+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRunMasked_SharedOutputKeepsOrder(t *testing.T) {
	out := tempFile(t, "both")
	runSh(t, `echo a; echo "$S" >&2; echo c`, out, out)
	if got, want := readFile(t, out), "a\n"+mask.Replacement+"\nc\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestRunMasked_ExitCode(t *testing.T) {
	if code := runSh(t, "exit 3", tempFile(t, "out"), tempFile(t, "err")); code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
}

func TestRunMasked_KilledBySignal(t *testing.T) {
	if code := runSh(t, "kill -TERM $$", tempFile(t, "out"), tempFile(t, "err")); code != 128+15 {
		t.Errorf("exit code = %d, want 143", code)
	}
}

func TestRunMasked_BackgroundProcessDoesNotKeepCredAlive(t *testing.T) {
	// The background sleep inherits the pipes; cred must still return soon
	// after the shell exits.
	stdout := tempFile(t, "out")
	start := time.Now()
	code := runSh(t, `sleep 5 & echo done`, stdout, tempFile(t, "err"))
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("returned after %v; the background process kept cred alive", elapsed)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got := readFile(t, stdout); got != "done\n" {
		t.Errorf("stdout = %q, want %q", got, "done\n")
	}
}

func TestRunMasked_StartError(t *testing.T) {
	_, err := runMasked(filepath.Join(t.TempDir(), "missing"), []string{"missing"}, nil, []string{testSecret},
		childIO{Stdout: tempFile(t, "out"), Stderr: tempFile(t, "err")})
	if err == nil {
		t.Fatal("want an error")
	}
}
