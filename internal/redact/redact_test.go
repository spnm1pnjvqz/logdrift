package redact

import (
	"testing"
)

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{"[invalid": "REDACTED"})
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestNew_ValidPatterns_NoError(t *testing.T) {
	_, err := New(map[string]string{`\d+`: "[NUM]"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoRules_ReturnsOriginal(t *testing.T) {
	r, _ := New(nil)
	got := r.Apply("hello world")
	if got != "hello world" {
		t.Fatalf("expected original string, got %q", got)
	}
}

func TestApply_SingleRule_ReplacesMatch(t *testing.T) {
	r, _ := New(map[string]string{`password=\S+`: "password=[REDACTED]"})
	input := "user login password=s3cr3t ok"
	want := "user login password=[REDACTED] ok"
	got := r.Apply(input)
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestApply_MultipleRules_AllApplied(t *testing.T) {
	r, _ := New(map[string]string{
		`token=\S+`:    "token=[REDACTED]",
		`\b\d{4}\b`: "[DIGITS]",
	})
	input := "token=abc123 code 1234"
	got := r.Apply(input)
	if got == input {
		t.Fatal("expected redaction to change the line")
	}
}

func TestApply_NoMatch_ReturnsOriginal(t *testing.T) {
	r, _ := New(map[string]string{`secret=\S+`: "[REDACTED]"})
	input := "nothing sensitive here"
	got := r.Apply(input)
	if got != input {
		t.Fatalf("expected unchanged line, got %q", got)
	}
}

func makeStringCh(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApplyToChannel_RedactsLines(t *testing.T) {
	r, _ := New(map[string]string{`\d+`: "[NUM]"})
	in := makeStringCh([]string{"line 123", "no digits", "val 99"})
	out := r.ApplyToChannel(in)

	results := []string{}
	for line := range out {
		results = append(results, line)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(results))
	}
	if results[0] != "line [NUM]" {
		t.Errorf("expected 'line [NUM]', got %q", results[0])
	}
	if results[1] != "no digits" {
		t.Errorf("expected 'no digits', got %q", results[1])
	}
	if results[2] != "val [NUM]" {
		t.Errorf("expected 'val [NUM]', got %q", results[2])
	}
}

func TestApplyToChannel_ClosesWhenInputClosed(t *testing.T) {
	r, _ := New(nil)
	in := makeStringCh(nil)
	out := r.ApplyToChannel(in)
	count := 0
	for range out {
		count++
	}
	if count != 0 {
		t.Fatalf("expected 0 lines, got %d", count)
	}
}
