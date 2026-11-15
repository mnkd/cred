// Package ref parses cred:// secret references.
package ref

import (
	"errors"
	"regexp"
	"strings"
)

// Prefix marks a value as a secret reference.
const Prefix = "cred://"

// ErrInvalid is returned for values that start with Prefix but are malformed.
var ErrInvalid = errors.New("invalid reference (expected cred://<namespace>/<KEY> using letters, digits, '_', '.' and '-')")

// validName matches a namespace or key name inside a reference.
var validName = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

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
// Values that only look like one ("cred:" or "cred//", e.g. a missing slash
// or colon) count too, so Parse rejects the typo instead of passing it on as
// a plain value. The scheme is case-sensitive.
func IsRef(value string) bool {
	return strings.HasPrefix(value, "cred:") || strings.HasPrefix(value, "cred//")
}

// Parse parses a cred://<namespace>/<KEY> reference.
func Parse(value string) (Ref, error) {
	if !strings.HasPrefix(value, Prefix) {
		return Ref{}, ErrInvalid
	}
	ns, key, ok := strings.Cut(strings.TrimPrefix(value, Prefix), "/")
	if !ok || !validName.MatchString(ns) || !validName.MatchString(key) {
		return Ref{}, ErrInvalid
	}
	return Ref{Namespace: ns, Key: key}, nil
}
