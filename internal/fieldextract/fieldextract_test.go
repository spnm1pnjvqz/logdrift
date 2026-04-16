package fieldextract_test

import (
	"testing"

	"github.com/user/logdrift/internal/fieldextract"
	"github.com/user/logdrift/internal/runner"
)

func TestNew_NoFields_ReturnsError(t *testing.T) {
	_, err := fieldextract.New(nil, false)
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestNew_ValidFields_NoError(t *testing.T) {
	_, err := fieldextract.New([]string{"level"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtract_KV_ReturnsRequestedFields(t *testing.T) {
	e, _ := fieldextract.New([]string{"level", "msg"}, false)
	got := e.Extract(`level=info msg="hello world" ignored=yes`)
	if got["level"] != "info" {
		t.Errorf("level: got %q", got["level"])
	}
	if got["msg"] != "hello world" {
		t.Errorf("msg: got %q", got["msg"])
	}
	if _, ok := got["ignored"]; ok {
		t.Error("ignored field should not be present")
	}
}

func TestExtract_KV_MissingField_NotInMap(t *testing.T) {
	e, _ := fieldextract.New([]string{"level"}, false)
	got := e.Extract("no fields here")
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestExtract_JSON_ReturnsRequestedFields(t *testing.T) {
	e, _ := fieldextract.New([]string{"level", "service"}, true)
	got := e.Extract(`{"level":"warn","service":"api","other":"x"}`)
	if got["level"] != "warn" {
		t.Errorf("level: got %q", got["level"])
	}
	if got["service"] != "api" {
		t.Errorf("service: got %q", got["service"])
	}
}

func TestExtract_JSON_InvalidJSON_ReturnsEmpty(t *testing.T) {
	e, _ := fieldextract.New([]string{"level"}, true)
	got := e.Extract("not json")
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_AnnotatesLines(t *testing.T) {
	e, _ := fieldextract.New([]string{"level"}, false)
	in := makeLineCh([]runner.LogLine{
		{Service: "svc", Text: "level=error something happened"},
		{Service: "svc", Text: "no fields"},
	})
	out := e.Apply(in)
	var results []runner.LogLine
	for l := range out {
		results = append(results, l)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if results[0].Text == "level=error something happened" {
		t.Error("expected annotation on first line")
	}
	if results[1].Text != "no fields" {
		t.Errorf("second line should be unchanged, got %q", results[1].Text)
	}
}
