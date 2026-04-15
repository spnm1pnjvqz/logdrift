package differ

import (
	"testing"
	"time"
)

func line(svc, text string) Line {
	return Line{Service: svc, Text: text, Timestamp: time.Now()}
}

func TestDiffer_NoneMode_NeverDrift(t *testing.T) {
	d := New(DiffModeNone, 0)
	l := line("svc", "hello world")
	if d.IsDrift(l) {
		t.Fatal("none mode should never report drift")
	}
}

func TestDiffer_UniqMode_FirstLineDrifts(t *testing.T) {
	d := New(DiffModeUniq, 0)
	l := line("a", "error: connection refused")
	if !d.IsDrift(l) {
		t.Fatal("unseen line should be drift")
	}
}

func TestDiffer_UniqMode_RecordedLineNotDrift(t *testing.T) {
	d := New(DiffModeUniq, 0)
	l := line("a", "info: starting up")
	d.Record(l)
	l2 := line("b", "info: starting up")
	if d.IsDrift(l2) {
		t.Fatal("seen line should not be drift")
	}
}

func TestDiffer_UniqMode_CaseInsensitive(t *testing.T) {
	d := New(DiffModeUniq, 0)
	d.Record(line("a", "INFO: ready"))
	if d.IsDrift(line("b", "info: ready")) {
		t.Fatal("normalisation should make these equal")
	}
}

func TestDiffer_FuzzyMode_HighSimilarityNotDrift(t *testing.T) {
	d := New(DiffModeFuzzy, 0.7)
	d.Record(line("a", "connected to database at 10.0.0.1"))
	// Very similar but different IP
	if d.IsDrift(line("b", "connected to database at 10.0.0.2")) {
		t.Fatal("similar lines should not be drift under fuzzy mode")
	}
}

func TestDiffer_FuzzyMode_DifferentLineIsDrift(t *testing.T) {
	d := New(DiffModeFuzzy, 0.8)
	d.Record(line("a", "server started on port 8080"))
	if !d.IsDrift(line("b", "panic: nil pointer dereference")) {
		t.Fatal("very different lines should be drift")
	}
}

func TestSimilarity_Identical(t *testing.T) {
	if similarity("hello", "hello") != 1.0 {
		t.Fatal("identical strings must have similarity 1.0")
	}
}

func TestSimilarity_Empty(t *testing.T) {
	if similarity("", "hello") != 0.0 {
		t.Fatal("empty string must have similarity 0.0")
	}
}
