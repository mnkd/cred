package mask

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// write feeds chunks to a Writer and returns what reached the underlying
// writer after each chunk, and the final output after Flush.
func write(t *testing.T, values []string, chunks ...string) (afterEach []string, final string) {
	t.Helper()
	var out bytes.Buffer
	w := NewWriter(&out, values)
	for _, c := range chunks {
		n, err := w.Write([]byte(c))
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if n != len(c) {
			t.Fatalf("Write returned %d, want %d", n, len(c))
		}
		afterEach = append(afterEach, out.String())
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return afterEach, out.String()
}

func TestWriter_MasksValue(t *testing.T) {
	_, got := write(t, []string{"s3cr3t-value"}, "token=s3cr3t-value\n")
	if want := "token=" + Replacement + "\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_MasksEveryOccurrence(t *testing.T) {
	_, got := write(t, []string{"s3cr3t-value"}, "s3cr3t-value s3cr3t-values3cr3t-value")
	want := Replacement + " " + Replacement + Replacement
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_ValueSplitAcrossChunks(t *testing.T) {
	_, got := write(t, []string{"s3cr3t-value"}, "a=s3c", "r3t-", "value\n")
	if want := "a=" + Replacement + "\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_HoldsBackOnlyPartialMatch(t *testing.T) {
	afterEach, _ := write(t, []string{"s3cr3t-value"}, "hello s3cr")
	if want := "hello "; afterEach[0] != want {
		t.Errorf("after first chunk got %q, want %q", afterEach[0], want)
	}
}

func TestWriter_PartialMatchThatDoesNotCompleteIsWritten(t *testing.T) {
	_, got := write(t, []string{"s3cr3t-value"}, "s3cr3t-", "other\n")
	if want := "s3cr3t-other\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_FlushWritesHeldBackTail(t *testing.T) {
	afterEach, got := write(t, []string{"s3cr3t-value"}, "end s3cr3t")
	if afterEach[0] != "end " {
		t.Errorf("after chunk got %q, want %q", afterEach[0], "end ")
	}
	if want := "end s3cr3t"; got != want {
		t.Errorf("after Flush got %q, want %q", got, want)
	}
}

func TestWriter_NewlineTerminatedLineIsWrittenAtOnce(t *testing.T) {
	line := `{"jsonrpc":"2.0","id":1,"result":{}}` + "\n"
	afterEach, _ := write(t, []string{"s3cr3t-value", "another-secret"}, line)
	if afterEach[0] != line {
		t.Errorf("got %q, want %q (nothing held back)", afterEach[0], line)
	}
}

func TestWriter_LongerValueWins(t *testing.T) {
	// "abcdefgh" is a prefix of the longer value.
	_, got := write(t, []string{"abcdefgh", "abcdefgh-longer"}, "x abcdefgh-longer y abcdefgh z\n")
	want := "x " + Replacement + " y " + Replacement + " z\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_LongerValueWinsAcrossChunks(t *testing.T) {
	// The shorter value is complete in the first chunk, but the longer one
	// may still follow, so it must not be replaced yet.
	_, got := write(t, []string{"abcdefgh", "abcdefgh-longer"}, "abcdefgh", "-longer\n")
	if want := Replacement + "\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_ShorterValueAtEndIsMaskedOnFlush(t *testing.T) {
	_, got := write(t, []string{"abcdefgh", "abcdefgh-longer"}, "abcdefgh")
	if got != Replacement {
		t.Errorf("got %q, want %q", got, Replacement)
	}
}

func TestWriter_ShortValuesAreNotMasked(t *testing.T) {
	_, got := write(t, []string{"1234567", ""}, "1234567\n")
	if want := "1234567\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_MinLengthValueIsMasked(t *testing.T) {
	_, got := write(t, []string{"12345678"}, "12345678\n")
	if want := Replacement + "\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_ManyChunksByteByByte(t *testing.T) {
	input := "k=s3cr3t-value; k2=other-secret-x\n"
	chunks := strings.Split(input, "")
	_, got := write(t, []string{"s3cr3t-value", "other-secret-x"}, chunks...)
	want := "k=" + Replacement + "; k2=" + Replacement + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValues(t *testing.T) {
	got := Values([]string{"short", "12345678", "abcdefgh-longer", "12345678", ""})
	want := []string{"abcdefgh-longer", "12345678"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_OverlapEndOfOneStartOfAnother(t *testing.T) {
	// The end of "abcdefgh" is the start of "defgh-xyz1".
	values := []string{"abcdefgh", "defgh-xyz1"}
	_, got := write(t, values, "[abcdefgh-xyz1]\n")
	if want := "[" + Replacement + "]\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_OverlapSplitRightAfterFirstValue(t *testing.T) {
	values := []string{"abcdefgh", "defgh-xyz1"}
	_, got := write(t, values, "[abcdefgh", "-xyz1]\n")
	if want := "[" + Replacement + "]\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriter_SuffixIsAlsoPrefix(t *testing.T) {
	// "abcabcab" ends with "abcab", which is also its start, so occurrences
	// in "abcabcabcab" overlap and form one run.
	values := []string{"abcabcab"}
	_, got := write(t, values, "x abcabcabcab y\n")
	if want := "x " + Replacement + " y\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	_, got = write(t, values, "x abcabc", "abcab y\n")
	if want := "x " + Replacement + " y\n"; got != want {
		t.Errorf("split: got %q, want %q", got, want)
	}
}

func TestWriter_MaskedJSONStaysValid(t *testing.T) {
	_, got := write(t, []string{"s3cr3t-value"}, `{"token":"s3cr3t-value"}`+"\n")
	if !json.Valid([]byte(got)) {
		t.Errorf("not valid JSON: %q", got)
	}
}
