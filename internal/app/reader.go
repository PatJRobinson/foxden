package app

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type renderedReaderLine struct {
	Text string
	Kind readerLineKind
}

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

	source := story.ContentSource
	if source == "" {
		source = "feed"
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
			b.WriteString(m.renderReaderLine(line))
			b.WriteString("\n")
		}
	}

	b.WriteString(mutedStyle.Render(fmt.Sprintf("content: %s", source)))

	if m.FetchingArticle {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("fetching article..."))
	}

	if story.ArticleFetchError != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("article fetch failed: " + story.ArticleFetchError))
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("f fetch article j/k scroll  ctrl+d/u page  g/G top/bottom  esc/q back"))

	return b.String()
}

func (m Model) renderReaderLine(line renderedReaderLine) string {
	switch line.Kind {
	case readerLineBlank:
		return ""
	case readerLineHeading:
		return subtitleStyle.Render(line.Text)
	case readerLineCode:
		return codeStyle.Render(line.Text)
	case readerLineList:
		return line.Text
	default:
		return line.Text
	}
}

func (m Model) readerLines() []renderedReaderLine {
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

	return formatReaderText(body, width)
}

func formatReaderText(text string, width int) []renderedReaderLine {
	rawLines := strings.Split(normalizeReaderInput(text), "\n")

	var out []renderedReaderLine
	lastWasBlank := true

	for _, rawLine := range rawLines {
		line := strings.TrimRight(rawLine, " \t")
		kind := classifyReaderLine(line)

		switch kind {
		case readerLineBlank:
			if !lastWasBlank {
				out = append(out, renderedReaderLine{Kind: readerLineBlank})
				lastWasBlank = true
			}

		case readerLineCode:
			out = append(out, renderedReaderLine{
				Text: line,
				Kind: readerLineCode,
			})
			lastWasBlank = false

		case readerLineHeading:
			if !lastWasBlank {
				out = append(out, renderedReaderLine{Kind: readerLineBlank})
			}
			out = append(out, renderedReaderLine{
				Text: strings.TrimSpace(line),
				Kind: readerLineHeading,
			})
			lastWasBlank = false

		case readerLineList:
			wrapped := wrapParagraphWithIndent(strings.TrimSpace(line), width, "  ")
			for _, wrappedLine := range wrapped {
				out = append(out, renderedReaderLine{
					Text: wrappedLine,
					Kind: readerLineList,
				})
			}
			lastWasBlank = false

		default:
			for _, wrappedLine := range wrapParagraph(strings.TrimSpace(line), width) {
				out = append(out, renderedReaderLine{
					Text: wrappedLine,
					Kind: readerLineProse,
				})
			}
			lastWasBlank = false
		}
	}

	return trimTrailingBlankReaderLines(out)
}

func normalizeReaderInput(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.TrimSpace(text)
}

func trimTrailingBlankReaderLines(lines []renderedReaderLine) []renderedReaderLine {
	for len(lines) > 0 && lines[len(lines)-1].Kind == readerLineBlank {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func wrapParagraphWithIndent(paragraph string, width int, continuationIndent string) []string {
	lines := wrapParagraph(paragraph, width)
	if len(lines) <= 1 {
		return lines
	}

	for i := 1; i < len(lines); i++ {
		lines[i] = continuationIndent + lines[i]
	}

	return lines
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
