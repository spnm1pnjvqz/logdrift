package bracket

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func makeLine(service, text string) runner.LogLine {
	return runner.LogLine{Service: service, Text: text}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_BothEmpty_ReturnsError(t *testing.T) {
	_, err := New("", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_OnlyOpen_NoError(t *testing.T) {
	_, err := New("[", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_OnlyClose_NoError(t *testing.T) {
	_, err := New("", "]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStamp_WrapsText(t *testing.T) {
	b, _ := New("[", "]")
	got := b.Stamp(makeLine("svc", "hello"))
	if got.Text != "[hello]" {
		t.Errorf("got %q, want %q", got.Text, "[hello]")
	}
}

func TestStamp_PreservesService(t *testing.T) {
	b, _ := New("(", ")")
	got := b.Stamp(makeLine("api", "ping"))
	if got.Service != "api" {
		t.Errorf("service changed: got %q", got.Service)
	}
}

func TestApply_WrapsAllLines(t *testing.T) {
	b, _ := New("<", ">")
	src := makeLineCh(
		makeLine("a", "foo"),
		makeLine("b", "bar"),
	)
	ctx := context.Background()
	var got []string
	for l := range b.Apply(ctx, src) {
		got = append(got, l.Text)
	}
	if len(got) != 2 || got[0] != "<foo>" || got[1] != "<bar>" {
		t.Errorf("unexpected output: %v", got)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	b, _ := New("[", "]")
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	out := b.Apply(ctx, ch)
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timed out waiting for channel close")
	}
}
