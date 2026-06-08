package app

import (
	"fmt"
	"strings"

	"github.com/patjrobinson/news-tui/internal/core"
)

func (m Model) View() string {
	if m.OverlayOpen {
		return m.renderOverlay()
	}
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	if m.Err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("error: %v", m.Err)))
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderBody())
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	base := b.String()

	return base
}

func (m Model) renderOverlay() string {
	text := m.OverlayText
	if strings.TrimSpace(text) == "" {
		text = "AI summary overlay\n\nNo summary text was generated.\n\n[esc/q close]"
	}

	width := m.Width - 8
	if width < 40 {
		width = 40
	}
	if width > 90 {
		width = 90
	}

	return overlayStyle.Width(width).Render(text)
}

func (m Model) renderHeader() string {
	topic := m.CurrentTopic()

	return titleStyle.Render(
		fmt.Sprintf("news-tui  topic=%s  range=%s", topic.Title, m.TimeRange),
	)
}

func (m Model) renderTabs() string {
	if len(m.Topics) == 0 {
		return mutedStyle.Render("no topics")
	}

	var parts []string

	for i, topic := range m.Topics {
		label := fmt.Sprintf("[%s]", topic.Title)
		if i == m.ActiveTopic {
			parts = append(parts, activeTabStyle.Render(label))
		} else {
			parts = append(parts, inactiveTabStyle.Render(label))
		}
	}

	return strings.Join(parts, " ")
}

func (m Model) renderBody() string {
	stories := m.CurrentStories()

	if len(stories) == 0 {
		return mutedStyle.Render("No stories for this topic yet.")
	}

	list := m.renderStoryList(stories)
	preview := m.renderPreview()

	return list + "\n\n" + preview
}

func (m Model) renderStoryList(stories []core.Story) string {
	var b strings.Builder

	visible := m.visibleStoryCount()
	if visible <= 0 {
		visible = 5
	}

	start := m.StoryOffset
	if start < 0 {
		start = 0
	}
	if start > len(stories) {
		start = len(stories)
	}

	end := start + visible
	if end > len(stories) {
		end = len(stories)
	}

	b.WriteString(fmt.Sprintf("Stories %d/%d\n", m.ActiveStory+1, len(stories)))

	if start > 0 {
		b.WriteString(mutedStyle.Render("  ↑ more"))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		story := stories[i]

		cursor := " "
		line := fmt.Sprintf("%s %s  %s", cursor, story.Title, mutedStyle.Render(story.SourceName))

		if i == m.ActiveStory {
			cursor = ">"
			line = fmt.Sprintf("%s %s  %s", cursor, story.Title, mutedStyle.Render(story.SourceName))
			line = selectedStoryStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	if end < len(stories) {
		b.WriteString(mutedStyle.Render("  ↓ more"))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderPreview() string {
	story, ok := m.CurrentStory()
	if !ok {
		return mutedStyle.Render("No story selected.")
	}

	var b strings.Builder

	b.WriteString("Preview\n")
	b.WriteString(boxStyle.Render(fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s",
		story.Title,
		story.Excerpt,
		mutedStyle.Render(story.URL),
		mutedStyle.Render(fmt.Sprintf("source: %s", story.SourceName)),
	)))

	return b.String()
}

func (m Model) renderFooter() string {
	leader := ""
	if m.LeaderPending {
		leader = " leader"
	}

	status := ""
	if m.LoadingCount > 0 {
		status = fmt.Sprintf("  loading %d...", m.LoadingCount)
	}

	return footerStyle.Render(
		fmt.Sprintf("h/l tabs  j/k stories  t range  ,+space summary soon  q quit%s%s", leader, status),
	)
}
