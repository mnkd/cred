// Package ref parses cred:// secret references.
package ref

import (
	"errors"
	"strings"
)

// Prefix marks a value as a secret reference.
const Prefix = "cred://"

// ErrInvalid is returned for values that start with Prefix but are malformed.
var ErrInvalid = errors.New("invalid reference (expected cred://<namespace>/<KEY>)")

// Ref points to a secret stored under a namespace and key.
type Ref struct {
	Namespace string
	Key       string
}

// String returns the reference in cred://<namespace>/<KEY> form.
func (r Ref) String() string {
	return Prefix + r.Namespace + "/" + r.Key
}

// IsRef reports whether value should be treated as a reference.
// The scheme is case-sensitive.
func IsRef(value string) bool {
	return strings.HasPrefix(value, Prefix)
}

// Parse parses a cred://<namespace>/<KEY> reference.
func Parse(value string) (Ref, error) {
	if !IsRef(value) {
		return Ref{}, ErrInvalid
	}
	ns, key, ok := strings.Cut(strings.TrimPrefix(value, Prefix), "/")
	if !ok || ns == "" || key == "" || strings.Contains(key, "/") {
		return Ref{}, ErrInvalid
	}
	return Ref{Namespace: ns, Key: key}, nil
}
