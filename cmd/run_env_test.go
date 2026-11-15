//go:build darwin

package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mnkd/cred/internal/keyring"
)

// stubSecrets replaces getSecret with an in-memory store keyed by
// "<service>/<key>" and returns a counter of lookups per entry.
func stubSecrets(t *testing.T, secrets map[string]string) map[string]int {
	t.Helper()
	calls := map[string]int{}
	orig := getSecret
	getSecret = func(service, key string) (string, error) {
		id := service + "/" + key
		calls[id]++
		v, ok := secrets[id]
		if !ok {
			return "", keyring.ErrNotFound
		}
		return v, nil
	}
	t.Cleanup(func() { getSecret = orig })
	return calls
}

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func envMap(environ []string) map[string]string {
	m := map[string]string{}
	for _, e := range environ {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}

func TestBuildRunEnv_ReferenceOnlySkipsNamespaceInjection(t *testing.T) {
	stubSecrets(t, map[string]string{
		"cred:work/JIRA_TOKEN": "jira-secret",
		"cred:default/OTHER":   "other-secret",
	})

	got, err := buildRunEnv(runEnvInput{
		Environ:       []string{"API_TOKEN=cred://work/JIRA_TOKEN"},
		Namespace:     "default",
		NamespaceKeys: []string{"OTHER"},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	m := envMap(got)
	if m["API_TOKEN"] != "jira-secret" {
		t.Errorf("API_TOKEN = %q, want jira-secret", m["API_TOKEN"])
	}
	if _, ok := m["OTHER"]; ok {
		t.Error("default namespace must not be injected when references are used without -n")
	}
}

func TestBuildRunEnv_NoReferencesInjectsNamespace(t *testing.T) {
	stubSecrets(t, map[string]string{"cred:default/OTHER": "other-secret"})

	got, err := buildRunEnv(runEnvInput{
		Environ:       []string{"PATH=/usr/bin"},
		Namespace:     "default",
		NamespaceKeys: []string{"OTHER"},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if envMap(got)["OTHER"] != "other-secret" {
		t.Error("namespace should be injected when there are no references (backward compatible)")
	}
}

func TestBuildRunEnv_EnvFileSkipsNamespaceInjection(t *testing.T) {
	stubSecrets(t, map[string]string{"cred:default/OTHER": "other-secret"})
	path := writeEnvFile(t, "PLAIN=value\n")

	got, err := buildRunEnv(runEnvInput{
		Namespace:     "default",
		NamespaceKeys: []string{"OTHER"},
		EnvFiles:      []string{path},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	m := envMap(got)
	if _, ok := m["OTHER"]; ok {
		t.Error("namespace must not be injected when --env-file is used without -n")
	}
	if m["PLAIN"] != "value" {
		t.Errorf("PLAIN = %q, want value", m["PLAIN"])
	}
}

func TestBuildRunEnv_ExplicitNamespaceWithReferences(t *testing.T) {
	stubSecrets(t, map[string]string{
		"cred:work/A":       "ref-a",
		"cred:prod/NS_ONLY": "ns-only",
	})

	got, err := buildRunEnv(runEnvInput{
		Environ:           []string{"A=cred://work/A"},
		Namespace:         "prod",
		NamespaceKeys:     []string{"NS_ONLY"},
		NamespaceExplicit: true,
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	m := envMap(got)
	if m["A"] != "ref-a" || m["NS_ONLY"] != "ns-only" {
		t.Errorf("got A=%q NS_ONLY=%q", m["A"], m["NS_ONLY"])
	}
}

func TestBuildRunEnv_Precedence(t *testing.T) {
	stubSecrets(t, map[string]string{
		"cred:prod/SHARED":   "from-namespace",
		"cred:work/FROM_REF": "from-env-file-ref",
	})
	first := writeEnvFile(t, "SHARED=cred://work/FROM_REF\nFILE_ONLY=first\nINHERITED=from-file\n")
	second := writeEnvFile(t, "FILE_ONLY=second\n")

	got, err := buildRunEnv(runEnvInput{
		Environ:           []string{"INHERITED=from-env", "SHARED=inherited"},
		Namespace:         "prod",
		NamespaceKeys:     []string{"SHARED"},
		NamespaceExplicit: true,
		EnvFiles:          []string{first, second},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	m := envMap(got)
	// env-file beats namespace injection, and its reference is resolved afterwards
	if m["SHARED"] != "from-env-file-ref" {
		t.Errorf("SHARED = %q, want from-env-file-ref", m["SHARED"])
	}
	// later env-file wins
	if m["FILE_ONLY"] != "second" {
		t.Errorf("FILE_ONLY = %q, want second", m["FILE_ONLY"])
	}
	// env-file beats inherited environment
	if m["INHERITED"] != "from-file" {
		t.Errorf("INHERITED = %q, want from-file", m["INHERITED"])
	}
}

func TestBuildRunEnv_NamespaceOverridesInheritedReference(t *testing.T) {
	// Layers are merged before resolution, so an overwritten reference is never looked up.
	calls := stubSecrets(t, map[string]string{"cred:prod/TOKEN": "ns-value"})

	got, err := buildRunEnv(runEnvInput{
		Environ:           []string{"TOKEN=cred://missing/TOKEN"},
		Namespace:         "prod",
		NamespaceKeys:     []string{"TOKEN"},
		NamespaceExplicit: true,
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if envMap(got)["TOKEN"] != "ns-value" {
		t.Errorf("TOKEN = %q, want ns-value", envMap(got)["TOKEN"])
	}
	if calls["cred:missing/TOKEN"] != 0 {
		t.Error("overwritten reference should not be resolved")
	}
}

func TestBuildRunEnv_ResolvesOnceAndCaches(t *testing.T) {
	calls := stubSecrets(t, map[string]string{"cred:work/T": "cred://work/OTHER"})

	got, err := buildRunEnv(runEnvInput{
		Environ: []string{"A=cred://work/T", "B=cred://work/T"},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	m := envMap(got)
	if m["A"] != "cred://work/OTHER" || m["B"] != "cred://work/OTHER" {
		t.Errorf("resolved values must not be resolved again: A=%q B=%q", m["A"], m["B"])
	}
	if calls["cred:work/T"] != 1 {
		t.Errorf("same reference fetched %d times, want 1", calls["cred:work/T"])
	}
}

func TestBuildRunEnv_ResolutionFailureReportsAll(t *testing.T) {
	stubSecrets(t, map[string]string{"cred:work/OK": "super-secret"})

	var stderr strings.Builder
	_, err := buildRunEnv(runEnvInput{
		Environ: []string{
			"OK=cred://work/OK",
			"MISSING=cred://work/NOPE",
			"BROKEN=cred://work",
		},
	}, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}

	out := stderr.String() + err.Error()
	for _, want := range []string{"MISSING (cred://work/NOPE)", "BROKEN (cred://work)"} {
		if !strings.Contains(out, want) {
			t.Errorf("output should mention %q: %s", want, out)
		}
	}
	if strings.Contains(out, "super-secret") {
		t.Errorf("output leaks a secret value: %s", out)
	}
}

func TestBuildRunEnv_EnvFileNotFound(t *testing.T) {
	stubSecrets(t, nil)
	_, err := buildRunEnv(runEnvInput{
		EnvFiles: []string{filepath.Join(t.TempDir(), "missing.env")},
	}, io.Discard)
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestBuildRunEnv_FailedReferenceLookedUpOnce(t *testing.T) {
	calls := stubSecrets(t, nil)

	_, err := buildRunEnv(runEnvInput{
		Environ: []string{"A=cred://work/NOPE", "B=cred://work/NOPE"},
	}, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if calls["cred:work/NOPE"] != 1 {
		t.Errorf("failing reference fetched %d times, want 1", calls["cred:work/NOPE"])
	}
}

func TestBuildRunEnv_NoSuccessMessagesOnFailure(t *testing.T) {
	stubSecrets(t, map[string]string{"cred:prod/NS_ONLY": "ns-only"})

	var stderr strings.Builder
	_, err := buildRunEnv(runEnvInput{
		Environ:           []string{"A=cred://work/NOPE"},
		Namespace:         "prod",
		NamespaceKeys:     []string{"NS_ONLY"},
		NamespaceExplicit: true,
	}, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(stderr.String(), "Injected") {
		t.Errorf("stderr should not report injection when the command is not started: %s", stderr.String())
	}
}
