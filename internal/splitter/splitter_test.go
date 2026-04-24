package splitter

import (
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func TestNew_NoRules_ReturnsError(t *testing.T) {
	_, err := New(nil, "")
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{"bad": `[invalid`}, "")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNew_EmptyBucketName_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{"": `foo`}, "")
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestNew_ValidRules_NoError(t *testing.T) {
	_, err := New(map[string]string{"errors": `error`}, "other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoute_MatchesFirstRule(t *testing.T) {
	s, _ := New(map[string]string{
		"errors": `(?i)error`,
		"warn":   `(?i)warn`,
	}, "other")

	if got := s.Route("ERROR: disk full"); got != "errors" {
		t.Errorf("expected errors, got %q", got)
	}
	if got := s.Route("WARN: low memory"); got != "warn" {
		t.Errorf("expected warn, got %q", got)
	}
}

func TestRoute_NoMatch_UsesDefault(t *testing.T) {
	s, _ := New(map[string]string{"errors": `error`}, "other")
	if got := s.Route("info: all good"); got != "other" {
		t.Errorf("expected other, got %q", got)
	}
}

func TestRoute_NoMatch_EmptyDefault_ReturnsEmpty(t *testing.T) {
	s, _ := New(map[string]string{"errors": `error`}, "")
	if got := s.Route("info: all good"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func makeLine(svc, text string) runner.LogLine {
	return runner.LogLine{Service: svc, Text: text}
}
