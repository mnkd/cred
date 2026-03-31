// Package keyring provides a timeout-wrapped interface to the OS keyring.
// Inspired by GitHub CLI's internal/keyring implementation.
package keyring

import (
	"errors"
	"time"

	gokeyring "github.com/zalando/go-keyring"
)

const (
	// servicePrefix is prepended to namespace to form the keyring service name.
	// e.g., "cred:default", "cred:work"
	servicePrefix = "cred"

	// timeout for keyring operations to prevent CLI hangs.
	timeout = 3 * time.Second
)

var (
	ErrNotFound = errors.New("secret not found in keyring")
	ErrTimeout  = errors.New("keyring operation timed out")
)

// ServiceName returns the keyring service name for a given namespace.
// Format: "cred:{namespace}" (e.g., "cred:default", "cred:myproject")
func ServiceName(namespace string) string {
	return servicePrefix + ":" + namespace
}

// Set stores a secret in the OS keyring with a timeout.
func Set(service, key, secret string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- gokeyring.Set(service, key, secret)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// Get retrieves a secret from the OS keyring with a timeout.
func Get(service, key string) (string, error) {
	type result struct {
		val string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		defer close(ch)
		val, err := gokeyring.Get(service, key)
		ch <- result{val, err}
	}()
	select {
	case res := <-ch:
		if errors.Is(res.err, gokeyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return res.val, res.err
	case <-time.After(timeout):
		return "", ErrTimeout
	}
}

// Delete removes a secret from the OS keyring with a timeout.
func Delete(service, key string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- gokeyring.Delete(service, key)
	}()
	select {
	case err := <-ch:
		if errors.Is(err, gokeyring.ErrNotFound) {
			return ErrNotFound
		}
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}
