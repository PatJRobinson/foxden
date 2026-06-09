package app

import (
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	lines := wrapText("alpha beta gamma delta", 10)

	if len(lines) == 0 {
		t.Fatal("wrapText() returned no lines")
	}

	for _, line := range lines {
		if len(line) > 10 {
			t.Fatalf("line %q length = %d, want <= 10", line, len(line))
		}
	}
}

func TestWrapTextPreservesParagraphBreaks(t *testing.T) {
	lines := wrapText("first paragraph\n\nsecond paragraph", 80)

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "first paragraph\n\nsecond paragraph") {
		t.Fatalf("wrapped text did not preserve paragraph break: %q", joined)
	}
}
