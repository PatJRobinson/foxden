package app

import (
	"strings"
	"unicode"
)

type readerLineKind int

const (
	readerLineBlank readerLineKind = iota
	readerLineProse
	readerLineHeading
	readerLineCode
	readerLineList
)

func classifyReaderLine(line string) readerLineKind {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return readerLineBlank
	}

	if isCodeLikeLine(trimmed) {
		return readerLineCode
	}

	if isListLine(trimmed) {
		return readerLineList
	}

	if isHeadingLikeLine(trimmed) {
		return readerLineHeading
	}

	return readerLineProse
}

func isCodeLikeLine(line string) bool {
	if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
		return true
	}

	prefixes := []string{
		"$ ",
		"# ",
		"sudo ",
		"go ",
		"git ",
		"nix ",
		"make",
		"sudo",
		"./",
		"curl ",
		"wget ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	if strings.Contains(line, "://") {
		return true
	}

	return false
}

func isListLine(line string) bool {
	prefixes := []string{"- ", "* ", "+ "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	if len(line) >= 3 && unicode.IsDigit(rune(line[0])) {
		for i, r := range line {
			if !unicode.IsDigit(r) {
				return r == '.' && i+1 < len(line) && line[i+1] == ' '
			}
		}
	}

	return false
}

func isHeadingLikeLine(line string) bool {
	if len(line) > 90 {
		return false
	}

	if strings.HasSuffix(line, ":") || strings.HasSuffix(line, "?") {
		return true
	}

	// Short title-ish line with no ending punctuation.
	if len(line) <= 60 &&
		!strings.HasSuffix(line, ".") &&
		!strings.HasSuffix(line, ",") &&
		!strings.HasSuffix(line, ";") {
		return true
	}

	return false
}
