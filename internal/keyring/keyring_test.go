package keyring

import (
	"testing"

	gokeyring "github.com/zalando/go-keyring"
)

func init() {
	// テスト用のモックバックエンドを使用（実際の OS キーリングにアクセスしない）
	gokeyring.MockInit()
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		namespace string
		want      string
	}{
		{"default", "cred:default"},
		{"work", "cred:work"},
		{"production", "cred:production"},
	}
	for _, tt := range tests {
		got := ServiceName(tt.namespace)
		if got != tt.want {
			t.Errorf("ServiceName(%q) = %q, want %q", tt.namespace, got, tt.want)
		}
	}
}

func TestSetAndGet(t *testing.T) {
	service := ServiceName("test")
	key := "TEST_KEY"
	secret := "test-secret-value"

	if err := Set(service, key, secret); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := Get(service, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != secret {
		t.Errorf("got %q, want %q", got, secret)
	}
}

func TestGet_NotFound(t *testing.T) {
	service := ServiceName("test-notfound")
	_, err := Get(service, "NONEXISTENT_KEY")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	service := ServiceName("test-delete")
	key := "DELETE_ME"

	if err := Set(service, key, "value"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := Delete(service, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := Get(service, key)
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound after delete", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	service := ServiceName("test-delete-notfound")
	err := Delete(service, "NONEXISTENT")
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSetOverwrite(t *testing.T) {
	service := ServiceName("test-overwrite")
	key := "OVERWRITE_KEY"

	if err := Set(service, key, "first"); err != nil {
		t.Fatalf("Set first: %v", err)
	}
	if err := Set(service, key, "second"); err != nil {
		t.Fatalf("Set second: %v", err)
	}

	got, err := Get(service, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}
