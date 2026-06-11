package app

import (
	"strings"
	"testing"
)

func TestClassifyReaderLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want readerLineKind
	}{
		{
			name: "blank",
			line: "   ",
			want: readerLineBlank,
		},
		{
			name: "heading ending colon",
			line: "How to build:",
			want: readerLineHeading,
		},
		{
			name: "heading ending question",
			line: "What does pi have to do with my data?",
			want: readerLineHeading,
		},
		{
			name: "sudo command",
			line: "sudo apt-get install automake",
			want: readerLineCode,
		},
		{
			name: "dot slash command",
			line: "./configure",
			want: readerLineCode,
		},
		{
			name: "url",
			line: "https://example.com/project",
			want: readerLineCode,
		},
		{
			name: "bullet list",
			line: "- first item",
			want: readerLineList,
		},
		{
			name: "numbered list",
			line: "1. first item",
			want: readerLineList,
		},
		{
			name: "prose",
			line: "This is a normal sentence with punctuation.",
			want: readerLineProse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyReaderLine(tt.line)
			if got != tt.want {
				t.Fatalf("classifyReaderLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestFormatReaderTextPreservesStructure(t *testing.T) {
	lines := formatReaderText(`Build instructions:
sudo apt-get install automake
./configure

- first long item that should wrap onto a continuation line when width is small
This is normal prose that should be wrapped.`, 38)

	if len(lines) == 0 {
		t.Fatal("formatReaderText() returned no lines")
	}

	if strings.TrimSpace(lines[0].Text) != "Build instructions:" || lines[0].Kind != readerLineHeading {
		t.Fatalf("line 0 = (%q, %v), want heading", lines[0].Text, lines[0].Kind)
	}

	if strings.TrimSpace(lines[1].Text) != "sudo apt-get install automake" || lines[1].Kind != readerLineCode {
		t.Fatalf("line 1 = (%q, %v), want code", lines[1].Text, lines[1].Kind)
	}

	if strings.TrimSpace(lines[2].Text) != "./configure" || lines[2].Kind != readerLineCode {
		t.Fatalf("line 2 = (%q, %v), want code", lines[2].Text, lines[2].Kind)
	}

	foundBlank := false
	foundList := false
	foundProse := false

	for _, line := range lines {
		switch line.Kind {
		case readerLineBlank:
			foundBlank = true
		case readerLineList:
			foundList = true
		case readerLineProse:
			foundProse = true
		}
	}

	if !foundBlank {
		t.Fatalf("formatReaderText() did not preserve a blank line: %#v", lines)
	}

	if !foundList {
		t.Fatalf("formatReaderText() did not produce a list line: %#v", lines)
	}

	if !foundProse {
		t.Fatalf("formatReaderText() did not produce a prose line: %#v", lines)
	}
}

func TestFormatReaderTextCollapsesRepeatedBlankLines(t *testing.T) {
	lines := formatReaderText("first\n\n\n\nsecond", 80)

	blankCount := 0
	for _, line := range lines {
		if line.Kind == readerLineBlank {
			blankCount++
		}
	}

	if blankCount != 1 {
		t.Fatalf("blank line count = %d, want 1; lines = %#v", blankCount, lines)
	}
}

func TestFormatReaderTextTrimsTrailingBlankLines(t *testing.T) {
	lines := formatReaderText("first\n\n", 80)

	if len(lines) == 0 {
		t.Fatal("formatReaderText() returned no lines")
	}

	last := lines[len(lines)-1]
	if last.Kind == readerLineBlank {
		t.Fatalf("last line is blank; lines = %#v", lines)
	}
}
