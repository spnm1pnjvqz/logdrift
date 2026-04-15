package highlight

import (
	"strings"
	"testing"
)

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{"[": "31"})
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestNew_ValidPatterns_NoError(t *testing.T) {
	h, err := New(map[string]string{"ERROR": "31", "WARN": "33"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Highlighter")
	}
}

func TestApply_NoRules_ReturnsOriginal(t *testing.T) {
	h, _ := New(map[string]string{})
	input := "hello world"
	if got := h.Apply(input); got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestApply_MatchingKeyword_WrapsWithANSI(t *testing.T) {
	h, _ := New(map[string]string{"ERROR": "31"})
	result := h.Apply("2024/01/01 ERROR something broke")
	if !strings.Contains(result, "\033[31m") {
		t.Fatal("expected ANSI red code in output")
	}
	if !strings.Contains(result, ansiReset) {
		t.Fatal("expected ANSI reset code in output")
	}
	if !strings.Contains(result, "ERROR") {
		t.Fatal("expected original keyword preserved")
	}
}

func TestApply_NoMatch_ReturnsUnchanged(t *testing.T) {
	h, _ := New(map[string]string{"ERROR": "31"})
	input := "everything is fine"
	if got := h.Apply(input); got != input {
		t.Fatalf("expected %q unchanged, got %q", input, got)
	}
}

func TestStripANSI_RemovesEscapes(t *testing.T) {
	h, _ := New(map[string]string{"WARN": "33"})
	coloured := h.Apply("WARN: disk space low")
	stripped := StripANSI(coloured)
	if strings.Contains(stripped, "\033") {
		t.Fatalf("expected no escape codes after strip, got %q", stripped)
	}
	if stripped != "WARN: disk space low" {
		t.Fatalf("expected original text, got %q", stripped)
	}
}

func TestApplyToLine_NilHighlighter_ReturnsOriginal(t *testing.T) {
	input := "some log line"
	if got := ApplyToLine(nil, input); got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestSummary_NoRules(t *testing.T) {
	h, _ := New(map[string]string{})
	if h.Summary() != "no highlight rules" {
		t.Fatalf("unexpected summary: %s", h.Summary())
	}
}

func TestSummary_WithRules(t *testing.T) {
	h, _ := New(map[string]string{"ERROR": "31"})
	s := h.Summary()
	if !strings.Contains(s, "1 rule(s)") {
		t.Fatalf("expected rule count in summary, got %q", s)
	}
}
