package filter_test

import (
	"testing"

	"github.com/yourorg/logdrift/internal/filter"
)

func TestNew_InvalidIncludePattern(t *testing.T) {
	_, err := filter.New(filter.Config{Include: []string{"[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid include pattern")
	}
}

func TestNew_InvalidExcludePattern(t *testing.T) {
	_, err := filter.New(filter.Config{Exclude: []string{"[bad"}})
	if err == nil {
		t.Fatal("expected error for invalid exclude pattern")
	}
}

func TestFilter_NoPatterns_AllowsAll(t *testing.T) {
	f, err := filter.New(filter.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"", "hello", "ERROR: boom"} {
		if !f.Allow(line) {
			t.Errorf("expected Allow(%q) = true", line)
		}
	}
}

func TestFilter_IncludeOnly(t *testing.T) {
	f, err := filter.New(filter.Config{Include: []string{"ERROR", "WARN"}})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Allow("ERROR: something") {
		t.Error("expected ERROR line to be allowed")
	}
	if !f.Allow("WARN: heads up") {
		t.Error("expected WARN line to be allowed")
	}
	if f.Allow("INFO: boring") {
		t.Error("expected INFO line to be blocked")
	}
}

func TestFilter_ExcludeOnly(t *testing.T) {
	f, err := filter.New(filter.Config{Exclude: []string{"DEBUG"}})
	if err != nil {
		t.Fatal(err)
	}
	if f.Allow("DEBUG: noisy") {
		t.Error("expected DEBUG line to be excluded")
	}
	if !f.Allow("INFO: useful") {
		t.Error("expected INFO line to be allowed")
	}
}

func TestFilter_IncludeAndExclude(t *testing.T) {
	f, err := filter.New(filter.Config{
		Include: []string{"ERROR"},
		Exclude: []string{"transient"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Allow("ERROR: real problem") {
		t.Error("expected non-transient ERROR to be allowed")
	}
	if f.Allow("ERROR: transient blip") {
		t.Error("expected transient ERROR to be excluded")
	}
	if f.Allow("INFO: unrelated") {
		t.Error("expected INFO line to be blocked by include filter")
	}
}
