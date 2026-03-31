package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	original := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	return dir, func() {
		if original == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", original)
		}
	}
}

func TestLoad_DefaultConfig(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("got version %d, want 1", cfg.Version)
	}
	if cfg.Default != "default" {
		t.Errorf("got default %q, want %q", cfg.Default, "default")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := &Config{
		Version:    1,
		Default:    "default",
		Namespaces: make(map[string]Namespace),
	}
	cfg.SetKeyMeta("default", "API_KEY", KeyMeta{
		Description: "test key",
		EnvVar:      "API_KEY",
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-01T00:00:00Z",
	})

	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	meta, ok := loaded.Namespaces["default"].Keys["API_KEY"]
	if !ok {
		t.Fatal("API_KEY not found after round-trip")
	}
	if meta.Description != "test key" {
		t.Errorf("got description %q, want %q", meta.Description, "test key")
	}
	if meta.EnvVar != "API_KEY" {
		t.Errorf("got env_var %q, want %q", meta.EnvVar, "API_KEY")
	}
}

func TestEnsureNamespace(t *testing.T) {
	cfg := defaultConfig()
	cfg.EnsureNamespace("work")

	if _, ok := cfg.Namespaces["work"]; !ok {
		t.Fatal("namespace 'work' not created")
	}

	// 二重呼び出しでも上書きされない
	cfg.SetKeyMeta("work", "KEY", KeyMeta{EnvVar: "KEY"})
	cfg.EnsureNamespace("work")
	if _, ok := cfg.Namespaces["work"].Keys["KEY"]; !ok {
		t.Fatal("existing key lost after second EnsureNamespace")
	}
}

func TestSetKeyMeta_And_DeleteKeyMeta(t *testing.T) {
	cfg := defaultConfig()
	cfg.SetKeyMeta("default", "FOO", KeyMeta{EnvVar: "FOO"})

	if _, ok := cfg.Namespaces["default"].Keys["FOO"]; !ok {
		t.Fatal("FOO not set")
	}

	cfg.DeleteKeyMeta("default", "FOO")
	if _, ok := cfg.Namespaces["default"].Keys["FOO"]; ok {
		t.Fatal("FOO should be deleted")
	}
}

func TestDeleteKeyMeta_NonexistentNamespace(t *testing.T) {
	cfg := defaultConfig()
	// パニックしないことを確認
	cfg.DeleteKeyMeta("nonexistent", "KEY")
}

func TestActiveNamespace(t *testing.T) {
	cfg := defaultConfig()

	// オーバーライドが優先
	if got := cfg.ActiveNamespace("work"); got != "work" {
		t.Errorf("got %q, want %q", got, "work")
	}

	// デフォルト
	if got := cfg.ActiveNamespace(""); got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}

	// config のデフォルト変更
	cfg.Default = "production"
	if got := cfg.ActiveNamespace(""); got != "production" {
		t.Errorf("got %q, want %q", got, "production")
	}
}

func TestConfigPath_XDG(t *testing.T) {
	dir := t.TempDir()
	original := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer func() {
		if original == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", original)
		}
	}()

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(dir, "cred", "config.yml")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg := defaultConfig()
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "cred", "config.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}
}
