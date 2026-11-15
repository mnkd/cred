//go:build integration

package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
)

var (
	credBin    string
	testcredBin string
)

// testNamespace is the namespace used for integration tests.
// Uses a distinct name to avoid colliding with real data.
const testNamespace = "cred-integration-test"

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "cred-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Build cred binary
	credBin = filepath.Join(tmpDir, "cred")
	out, err := exec.Command("go", "build", "-o", credBin, "..").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build cred: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Build testcred binary
	testcredBin = filepath.Join(tmpDir, "testcred")
	out, err = exec.Command("go", "build", "-o", testcredBin, "../cmd/testcred").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build testcred: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// setupConfig creates a temporary config directory with the given config
// and returns a cleanup function.
func setupConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()

	tmpDir := t.TempDir()
	credDir := filepath.Join(tmpDir, "cred")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(credDir, "config.yml"), data, 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return tmpDir
}

// setKeyring stores a secret and registers cleanup to delete it.
func setKeyring(t *testing.T, namespace, key, value string) {
	t.Helper()
	service := keyring.ServiceName(namespace)
	if err := keyring.Set(service, key, value); err != nil {
		t.Fatalf("failed to set keyring %s/%s: %v", namespace, key, err)
	}
	t.Cleanup(func() {
		keyring.Delete(service, key)
	})
}

// runCred executes the cred binary with the given args and config dir.
// Returns stdout, stderr, and any error.
func runCred(configDir string, args ...string) (stdout, stderr string, err error) {
	return runCredWithEnv(configDir, nil, args...)
}

// runCredWithEnv is runCred with extra KEY=VALUE entries in cred's environment.
func runCredWithEnv(configDir string, extraEnv []string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(credBin, args...)
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir)
	cmd.Env = append(cmd.Env, extraEnv...)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// parseOutput parses testcred output ("KEY=VALUE\n") into a map.
func parseOutput(s string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if line == "" {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		m[k] = v
	}
	return m
}

// TestSingleKeyInjection verifies:
// - A single keyring secret is injected as an environment variable into the child process
// - stderr reports the correct injection count ("Injected 1 credential(s)")
func TestSingleKeyInjection(t *testing.T) {
	const key = "TEST_SINGLE_KEY"
	const secret = "single-secret-value"

	cfg := &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: map[string]config.KeyMeta{
				key: {},
			}},
		},
	}

	configDir := setupConfig(t, cfg)
	setKeyring(t, testNamespace, key, secret)

	stdout, stderr, err := runCred(configDir, "run", "--", testcredBin, key)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	got := parseOutput(stdout)
	if got[key] != secret {
		t.Errorf("got %q, want %q", got[key], secret)
	}

	if !strings.Contains(stderr, "Injected 1 credential(s)") {
		t.Errorf("unexpected stderr: %s", stderr)
	}
}

// TestMultipleKeysInjection verifies:
// - All keys registered in the namespace are injected into the child process
// - Each key holds its correct secret value
// - stderr reports the correct total injection count
func TestMultipleKeysInjection(t *testing.T) {
	keys := map[string]string{
		"TEST_MULTI_A": "value-a",
		"TEST_MULTI_B": "value-b",
		"TEST_MULTI_C": "value-c",
	}

	keysConfig := make(map[string]config.KeyMeta)
	var keyNames []string
	for k := range keys {
		keysConfig[k] = config.KeyMeta{}
		keyNames = append(keyNames, k)
	}

	cfg := &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: keysConfig},
		},
	}

	configDir := setupConfig(t, cfg)
	for k, v := range keys {
		setKeyring(t, testNamespace, k, v)
	}

	args := append([]string{"run", "--", testcredBin}, keyNames...)
	stdout, stderr, err := runCred(configDir, args...)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	got := parseOutput(stdout)
	for k, want := range keys {
		if got[k] != want {
			t.Errorf("%s: got %q, want %q", k, got[k], want)
		}
	}

	if !strings.Contains(stderr, fmt.Sprintf("Injected %d credential(s)", len(keys))) {
		t.Errorf("unexpected stderr: %s", stderr)
	}
}

// TestNamespaceSelection verifies:
// - The -n flag selects a non-default namespace
// - Only keys from the selected namespace are injected (not from the default)
// - stderr mentions the selected namespace name
func TestNamespaceSelection(t *testing.T) {
	const ns = "cred-integration-test-alt"
	const key = "TEST_NS_KEY"
	const secret = "ns-secret"

	cfg := &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: map[string]config.KeyMeta{
				"TEST_DEFAULT_KEY": {},
			}},
			ns: {Keys: map[string]config.KeyMeta{
				key: {},
			}},
		},
	}

	configDir := setupConfig(t, cfg)
	setKeyring(t, ns, key, secret)

	stdout, stderr, err := runCred(configDir, "run", "-n", ns, "--", testcredBin, key)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	got := parseOutput(stdout)
	if got[key] != secret {
		t.Errorf("got %q, want %q", got[key], secret)
	}

	if !strings.Contains(stderr, fmt.Sprintf("namespace %q", ns)) {
		t.Errorf("stderr should mention namespace %q: %s", ns, stderr)
	}
}

// TestMissingKeyInKeyring verifies:
// - A key registered in config but absent from keyring is skipped with a warning
// - Other keys present in keyring are still injected normally
// - stderr contains "Skipping <key>" for the missing key
// - The injection count reflects only the successfully retrieved keys
func TestMissingKeyInKeyring(t *testing.T) {
	const presentKey = "TEST_PRESENT_KEY"
	const presentSecret = "present-value"
	const missingKey = "TEST_MISSING_KEY"

	cfg := &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: map[string]config.KeyMeta{
				presentKey: {},
				missingKey: {},
			}},
		},
	}

	configDir := setupConfig(t, cfg)
	// Only set presentKey; missingKey is not in the keyring
	setKeyring(t, testNamespace, presentKey, presentSecret)

	stdout, stderr, err := runCred(configDir, "run", "--", testcredBin, presentKey, missingKey)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	got := parseOutput(stdout)
	if got[presentKey] != presentSecret {
		t.Errorf("%s: got %q, want %q", presentKey, got[presentKey], presentSecret)
	}
	if got[missingKey] != "" {
		t.Errorf("%s: got %q, want empty", missingKey, got[missingKey])
	}

	if !strings.Contains(stderr, fmt.Sprintf("Skipping %s", missingKey)) {
		t.Errorf("stderr should warn about missing key: %s", stderr)
	}
	if !strings.Contains(stderr, "Injected 1 credential(s)") {
		t.Errorf("should inject only 1 credential: %s", stderr)
	}
}

// TestEmptyNamespace verifies:
// - A namespace with no keys still allows the command to run successfully
// - No environment variables are injected
// - stderr contains "No credentials in namespace" warning
func TestEmptyNamespace(t *testing.T) {
	const emptyNs = "cred-integration-test-empty"

	cfg := &config.Config{
		Version: 1,
		Default: emptyNs,
		Namespaces: map[string]config.Namespace{
			emptyNs: {Keys: map[string]config.KeyMeta{}},
		},
	}

	configDir := setupConfig(t, cfg)

	stdout, stderr, err := runCred(configDir, "run", "--", testcredBin, "ANYTHING")
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	// Command should still run
	got := parseOutput(stdout)
	if got["ANYTHING"] != "" {
		t.Errorf("ANYTHING should be empty, got %q", got["ANYTHING"])
	}

	if !strings.Contains(stderr, "No credentials in namespace") {
		t.Errorf("stderr should warn about empty namespace: %s", stderr)
	}
}

// TestEnvOverride verifies:
// - A pre-existing environment variable is overwritten by the keyring value
// - The child process sees the keyring value, not the original
func TestEnvOverride(t *testing.T) {
	const key = "TEST_OVERRIDE_KEY"
	const original = "original-value"
	const secret = "keyring-value"

	cfg := &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: map[string]config.KeyMeta{
				key: {},
			}},
		},
	}

	configDir := setupConfig(t, cfg)
	setKeyring(t, testNamespace, key, secret)

	// Run with the key already set in the environment
	cmd := exec.Command(credBin, "run", "--", testcredBin, key)
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+configDir,
		key+"="+original,
	)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, errBuf.String())
	}

	got := parseOutput(outBuf.String())
	if got[key] != secret {
		t.Errorf("got %q, want %q (keyring should override existing env)", got[key], secret)
	}
}

// refConfig returns a config whose default namespace holds otherKey, so tests
// can check that references alone do not inject the default namespace.
func refConfig(otherKey string) *config.Config {
	return &config.Config{
		Version: 1,
		Default: testNamespace,
		Namespaces: map[string]config.Namespace{
			testNamespace: {Keys: map[string]config.KeyMeta{
				otherKey: {},
			}},
		},
	}
}

// TestReferenceResolution verifies:
// - A cred:// reference in the environment is replaced with the keyring value
// - The variable name may differ from the keyring key
// - Without -n, the default namespace is not injected
// - stdout contains only the command's output (cred writes to stderr)
func TestReferenceResolution(t *testing.T) {
	const refKey = "TEST_REF_KEY"
	const refSecret = "ref-secret-value"
	const otherKey = "TEST_REF_OTHER_KEY"

	configDir := setupConfig(t, refConfig(otherKey))
	setKeyring(t, testNamespace, refKey, refSecret)
	setKeyring(t, testNamespace, otherKey, "other-value")

	extra := []string{"API_TOKEN=cred://" + testNamespace + "/" + refKey}
	stdout, stderr, err := runCredWithEnv(configDir, extra, "run", "--", testcredBin, "API_TOKEN", otherKey)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	want := "API_TOKEN=" + refSecret + "\n" + otherKey + "=\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if !strings.Contains(stderr, "Resolved 1 reference(s)") {
		t.Errorf("stderr should report resolution: %s", stderr)
	}
	if strings.Contains(stderr, refSecret) {
		t.Errorf("stderr leaks the secret: %s", stderr)
	}
}

// TestReferenceFailureDoesNotRun verifies:
// - An unresolvable reference stops cred with exit code 1
// - The command is not started (no stdout)
// - stderr names the variable and the reference
func TestReferenceFailureDoesNotRun(t *testing.T) {
	configDir := setupConfig(t, refConfig("TEST_REF_UNUSED"))

	ref := "cred://" + testNamespace + "/TEST_REF_DOES_NOT_EXIST"
	stdout, stderr, err := runCredWithEnv(configDir, []string{"API_TOKEN=" + ref}, "run", "--", testcredBin, "API_TOKEN")

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("want exit code 1, got %v\nstderr: %s", err, stderr)
	}
	if stdout != "" {
		t.Errorf("command should not run, got stdout %q", stdout)
	}
	if !strings.Contains(stderr, "API_TOKEN") || !strings.Contains(stderr, ref) {
		t.Errorf("stderr should name the variable and reference: %s", stderr)
	}
}

// TestEnvFileReference verifies:
// - --env-file values are passed through, and references in them are resolved
// - Without -n, the default namespace is not injected
func TestEnvFileReference(t *testing.T) {
	const refKey = "TEST_ENVFILE_REF_KEY"
	const refSecret = "envfile-secret-value"
	const otherKey = "TEST_ENVFILE_OTHER_KEY"

	configDir := setupConfig(t, refConfig(otherKey))
	setKeyring(t, testNamespace, refKey, refSecret)
	setKeyring(t, testNamespace, otherKey, "other-value")

	envFile := filepath.Join(t.TempDir(), ".env.cred")
	content := "API_TOKEN=\"cred://" + testNamespace + "/" + refKey + "\"\nPLAIN=plain-value\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runCred(configDir, "run", "--env-file", envFile, "--", testcredBin, "API_TOKEN", "PLAIN", otherKey)
	if err != nil {
		t.Fatalf("cred run failed: %v\nstderr: %s", err, stderr)
	}

	got := parseOutput(stdout)
	if got["API_TOKEN"] != refSecret {
		t.Errorf("API_TOKEN = %q, want %q", got["API_TOKEN"], refSecret)
	}
	if got["PLAIN"] != "plain-value" {
		t.Errorf("PLAIN = %q, want plain-value", got["PLAIN"])
	}
	if got[otherKey] != "" {
		t.Errorf("%s should not be injected, got %q", otherKey, got[otherKey])
	}
}
