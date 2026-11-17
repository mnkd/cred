// Package mask replaces known secret values in a byte stream.
package mask

import (
	"bytes"
	"cmp"
	"io"
	"slices"
)

// Replacement is written in place of each run of masked bytes.
const Replacement = "<concealed by cred>"

// MinLen is the shortest value that is masked. Masking a few common bytes
// would corrupt unrelated output.
const MinLen = 8

// Values returns the values that are masked: at least MinLen bytes long,
// without duplicates, longest first.
func Values(values []string) []string {
	var out []string
	for _, v := range values {
		if len(v) >= MinLen && !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	slices.SortStableFunc(out, func(a, b string) int { return cmp.Compare(len(b), len(a)) })
	return out
}

// Writer masks values written through it and passes the rest on at once.
// Only a tail that may still become part of a value is held back until the
// next Write or Flush. A Writer is not safe for concurrent use.
type Writer struct {
	w       io.Writer
	values  [][]byte
	pending []byte
}

// NewWriter returns a Writer that masks values (filtered by Values) on w.
func NewWriter(w io.Writer, values []string) *Writer {
	mw := &Writer{w: w}
	for _, v := range Values(values) {
		mw.values = append(mw.values, []byte(v))
	}
	return mw
}

// Write masks p and writes everything that can no longer change. It returns
// len(p) on success, including bytes that are held back.
func (m *Writer) Write(p []byte) (int, error) {
	if len(m.values) == 0 {
		return m.w.Write(p)
	}
	m.pending = append(m.pending, p...)
	if err := m.emit(false); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Flush masks and writes the held-back tail. Call it once the input ends.
func (m *Writer) Flush() error {
	return m.emit(true)
}

type span struct{ start, end int }

// emit writes the committed part of m.pending and keeps the rest. With final,
// no more input can arrive, so everything is committed.
func (m *Writer) emit(final bool) error {
	buf := m.pending
	if len(buf) == 0 {
		return nil
	}
	runs := m.runs(buf)

	commit := len(buf)
	if !final {
		commit = m.openPartial(buf)
		// A run that crosses the commit point may still grow: a value starting
		// at the commit point could overlap it once more bytes arrive.
		for _, r := range runs {
			if r.start < commit && commit < r.end {
				commit = r.start
				break
			}
		}
	}

	var out []byte
	pos := 0
	for _, r := range runs {
		if r.start >= commit {
			break
		}
		out = append(out, buf[pos:r.start]...)
		out = append(out, Replacement...)
		pos = r.end
	}
	out = append(out, buf[pos:commit]...)

	m.pending = append(m.pending[:0], buf[commit:]...)
	if len(out) == 0 {
		return nil
	}
	_, err := m.w.Write(out)
	return err
}

// runs returns every occurrence of every value in buf, merged into sorted
// runs. Overlapping occurrences merge; occurrences that only touch do not.
func (m *Writer) runs(buf []byte) []span {
	var spans []span
	for _, v := range m.values {
		for i := 0; ; {
			j := bytes.Index(buf[i:], v)
			if j < 0 {
				break
			}
			spans = append(spans, span{i + j, i + j + len(v)})
			i += j + 1
		}
	}
	slices.SortFunc(spans, func(a, b span) int { return cmp.Compare(a.start, b.start) })

	var merged []span
	for _, s := range spans {
		if n := len(merged); n > 0 && s.start < merged[n-1].end {
			merged[n-1].end = max(merged[n-1].end, s.end)
			continue
		}
		merged = append(merged, s)
	}
	return merged
}

// openPartial returns the earliest position from which the rest of buf is a
// proper prefix of some value, or len(buf) if there is none.
func (m *Writer) openPartial(buf []byte) int {
	earliest := len(buf)
	for _, v := range m.values {
		for i := max(0, len(buf)-len(v)+1); i < earliest; i++ {
			k := bytes.IndexByte(buf[i:earliest], v[0])
			if k < 0 {
				break
			}
			i += k
			if bytes.HasPrefix(v, buf[i:]) {
				earliest = i
				break
			}
		}
	}
	return earliest
}
