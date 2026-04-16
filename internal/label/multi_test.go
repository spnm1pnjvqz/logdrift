package label

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func TestLabelAll_NoSources_ReturnsError(t *testing.T) {
	_, err := LabelAll(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestLabelAll_EmptyServiceName_ReturnsError(t *testing.T) {
	sources := []ServiceChannel{{Service: "", Ch: make(chan runner.LogLine)}}
	_, err := LabelAll(context.Background(), sources)
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestLabelAll_LabelsEachChannel(t *testing.T) {
	lines := func(svc string) <-chan runner.LogLine {
		ch := make(chan runner.LogLine, 1)
		ch <- runner.LogLine{Text: "msg"}
		close(ch)
		return ch
	}
	sources := []ServiceChannel{
		{Service: "svc-a", Ch: lines("svc-a")},
		{Service: "svc-b", Ch: lines("svc-b")},
	}
	ctx := context.Background()
	outs, err := LabelAll(ctx, sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outs) != 2 {
		t.Fatalf("expected 2 output channels, got %d", len(outs))
	}
	expected := []string{"svc-a", "svc-b"}
	for i, ch := range outs {
		for l := range ch {
			if l.Service != expected[i] {
				t.Errorf("ch[%d]: expected service=%q, got %q", i, expected[i], l.Service)
			}
		}
	}
}
