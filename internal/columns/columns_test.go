package columns_test

import (
	"testing"

	"github.com/user/logdrift/internal/columns"
	"github.com/user/logdrift/internal/runner"
)

func makeLine(text string) runner.LogLine {
	return runner.LogLine{Service: "svc", Text: text}
}

func makeLineCh(lines ...string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- makeLine(l)
	}
	close(ch)
	return ch
}

func TestNew_EmptyDelimiter_ReturnsError(t *testing.T) {
	_, err := columns.New("", []int{10})
	if err == nil {
		t.Fatal("expected error for empty delimiter")
	}
}

func TestNew_NoWidths_ReturnsError(t *testing.T) {
	_, err := columns.New(" ", nil)
	if err == nil {
		t.Fatal("expected error for no widths")
	}
}

func TestNew_ZeroWidth_ReturnsError(t *testing.T) {
	_, err := columns.New(" ", []int{10, 0})
	if err == nil {
		t.Fatal("expected error for zero width")
	}
}

func TestNew_Valid_NoError(t *testing.T) {
	_, err := columns.New(" ", []int{10, 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormat_PadsFields(t *testing.T) {
	f, _ := columns.New(" ", []int{10, 15})
	got := f.Format(makeLine("INFO starting"))
	// first field "INFO" padded to 10, second "starting" padded to 15
	want := fmt.Sprintf("%-10s%-15s", "INFO", "starting")
	if got.Text != want {
		t.Fatalf("want %q got %q", want, got.Text)
	}
}

func TestFormat_ExtraFieldsAppended(t *testing.T) {
	f, _ := columns.New("|", []int{5})
	got := f.Format(makeLine("A|B|C"))
	if got.Text == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestApply_FormatsAllLines(t *testing.T) {
	f, _ := columns.New(" ", []int{8})
	src := makeLineCh("INFO msg", "WARN other")
	out := f.Apply(src)
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestApply_ClosesWhenSourceClosed(t *testing.T) {
	f, _ := columns.New(" ", []int{5})
	src := make(chan runner.LogLine)
	close(src)
	out := f.Apply(src)
	_, open := <-out
	if open {
		t.Fatal("expected output channel to be closed")
	}
}
