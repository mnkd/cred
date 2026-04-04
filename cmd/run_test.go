//go:build darwin

package cmd

import (
	"testing"

	"github.com/mnkd/cred/internal/config"
)

func TestSortedKeys(t *testing.T) {
	ns := config.Namespace{
		Keys: map[string]config.KeyMeta{
			"ZEBRA_KEY":  {},
			"ALPHA_KEY":  {},
			"MIDDLE_KEY": {},
		},
	}

	got := sortedKeys(ns.Keys)

	want := []string{"ALPHA_KEY", "MIDDLE_KEY", "ZEBRA_KEY"}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	ns := config.Namespace{
		Keys: map[string]config.KeyMeta{},
	}

	got := sortedKeys(ns.Keys)
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}
