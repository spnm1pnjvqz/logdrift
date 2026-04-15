package timestamp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/timestamp"
)

func makeLine(text string) runner.LogLine {
	return runner.LogLine{Service: "svc", Text: text, At: time.Now()}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_UnknownFormat_ReturnsError(t *testing.T) {
	_, err := timestamp.New("bogus")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestNew_ValidFormats_NoError(t *testing.T) {
	formats := []timestamp.Format{
		timestamp.FormatRFC3339,
		timestamp.FormatUnix,
		timestamp.FormatKitchen,
		timestamp.FormatRelative,
	}
	for _, f := range formats {
		_, err := timestamp.New(f)
		if err != nil {
			t.Errorf("format %q: unexpected error: %v", f, err)
		}
	}
}

func TestStamp_RFC3339_PrependedToText(t *testing.T) {
	s, _ := timestamp.New(timestamp.FormatRFC3339)
	original := makeLine("hello world")
	stamped := s.Stamp(original)
	if !strings.HasSuffix(stamped.Text, "hello world") {
		t.Errorf("expected suffix 'hello world', got %q", stamped.Text)
	}
	parts := strings.SplitN(stamped.Text, " ", 2)
	if _, err := time.Parse(time.RFC3339, parts[0]); err != nil {
		t.Errorf("prefix is not RFC3339: %q", parts[0])
	}
}

func TestStamp_Unix_NumericPrefix(t *testing.T) {
	s, _ := timestamp.New(timestamp.FormatUnix)
	stamped := s.Stamp(makeLine("msg"))
	parts := strings.SplitN(stamped.Text, " ", 2)
	if len(parts[0]) == 0 || parts[0][0] < '0' || parts[0][0] > '9' {
		t.Errorf("expected numeric unix prefix, got %q", parts[0])
	}
}

func TestStamp_Relative_StartsWithPlus(t *testing.T) {
	s, _ := timestamp.New(timestamp.FormatRelative)
	stamped := s.Stamp(makeLine("msg"))
	if !strings.HasPrefix(stamped.Text, "+") {
		t.Errorf("expected relative prefix starting with '+', got %q", stamped.Text)
	}
}

func TestApply_StampsAllLines(t *testing.T) {
	s, _ := timestamp.New(timestamp.FormatKitchen)
	lines := []runner.LogLine{makeLine("a"), makeLine("b"), makeLine("c")}
	out := s.Apply(makeLineCh(lines...))
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	for _, l := range got {
		if !.Contains(l.Text, ":") {
			t.Errorf("kitchen timestamp not found in %q", l.Text)
		}
	}
}

func TestApply_ClosesOutputWhenInputClosed(t *testing.T) {
	s, _ := timestamp.New(timestamp.FormatRFC3339)
	out := s.Apply(makeLineCh())
	_, open := <-out
	if open {
		t.Fatal("expected output channel to be closed")
	}
}
