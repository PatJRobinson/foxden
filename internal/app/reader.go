package app

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (m Model) renderReader() string {
	story := m.ReaderStory

	lines := m.readerLines()
	pageSize := m.readerPageSize()

	start := m.ReaderOffset
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}

	end := start + pageSize
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(story.Title))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%s  %s", story.SourceName, story.URL)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("line %d/%d", start+1, max(1, len(lines)))))
	b.WriteString("\n\n")

	if len(lines) == 0 {
		b.WriteString(mutedStyle.Render("No stored content for this story yet."))
		b.WriteString("\n")
	} else {
		for _, line := range lines[start:end] {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("j/k scroll  ctrl+d/u page  g/G top/bottom  esc/q back"))

	return b.String()
}

func (m Model) readerLines() []string {
	story := m.ReaderStory

	body := strings.TrimSpace(story.Content)
	if body == "" {
		body = strings.TrimSpace(story.Excerpt)
	}

	if body == "" {
		return nil
	}

	width := m.Width - 4
	if width < 40 {
		width = 40
	}
	if width > 100 {
		width = 100
	}

	return wrapText(body, width)
}

func wrapText(text string, width int) []string {
	var lines []string

	paragraphs := strings.Split(text, "\n")
	lastWasBlank := false

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			if !lastWasBlank && len(lines) > 0 {
				lines = append(lines, "")
				lastWasBlank = true
			}
			continue
		}

		lines = append(lines, wrapParagraph(paragraph, width)...)
		lastWasBlank = false
	}

	return lines
}

func wrapParagraph(paragraph string, width int) []string {
	words := strings.Fields(paragraph)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var current strings.Builder

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		currentLen := utf8.RuneCountInString(current.String())

		if currentLen == 0 {
			current.WriteString(word)
			continue
		}

		if currentLen+1+wordLen > width {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			continue
		}

		current.WriteString(" ")
		current.WriteString(word)
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) scrollReaderDown(n int) {
	maxOffset := m.readerMaxOffset()
	m.ReaderOffset += n
	if m.ReaderOffset > maxOffset {
		m.ReaderOffset = maxOffset
	}
}

func (m *Model) scrollReaderUp(n int) {
	m.ReaderOffset -= n
	if m.ReaderOffset < 0 {
		m.ReaderOffset = 0
	}
}

func (m Model) readerPageSize() int {
	size := m.Height - 8
	if size < 5 {
		return 5
	}
	return size
}

func (m Model) readerMaxOffset() int {
	lines := m.readerLines()
	pageSize := m.readerPageSize()

	maxOffset := len(lines) - pageSize
	if maxOffset < 0 {
		return 0
	}

	return maxOffset
}
