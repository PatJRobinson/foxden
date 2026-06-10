package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case topicStoriesLoadedMsg:
		if m.LoadingCount > 0 {
			m.LoadingCount--
		}

		m.Stories[msg.TopicID] = msg.Stories
		m.Err = nil

		if m.ActiveStory >= len(m.CurrentStories()) {
			m.ActiveStory = 0
			m.StoryOffset = 0
		}

		return m, nil

	case topicRefreshFailedMsg:
		if m.LoadingCount > 0 {
			m.LoadingCount--
		}

		m.Err = msg.Err
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case articleFetchedMsg:
		m.FetchingArticle = false
		m.Err = nil
		m.updateStory(msg.Story)

		if m.ViewMode == ViewModeReader && m.ReaderStory.ID == msg.Story.ID {
			m.ReaderStory = msg.Story
			m.ReaderOffset = 0
		}

		return m, nil

	case articleFetchFailedMsg:
		m.FetchingArticle = false
		m.Err = msg.Err
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ViewMode == ViewModeReader {
		switch msg.String() {
		case "esc", "q":
			m = m.CloseReader()
			return m, nil

		case "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			m.scrollReaderDown(1)
			return m, nil

		case "k", "up":
			m.scrollReaderUp(1)
			return m, nil

		case "ctrl+d", "pgdown":
			m.scrollReaderDown(m.readerPageSize())
			return m, nil

		case "ctrl+u", "pgup":
			m.scrollReaderUp(m.readerPageSize())
			return m, nil

		case "g":
			m.ReaderOffset = 0
			return m, nil

		case "G":
			m.ReaderOffset = m.readerMaxOffset()
			return m, nil

		case "f":
			story := m.ReaderStory
			if story.ID == "" {
				return m, nil
			}
			m.FetchingArticle = true
			m.Err = nil
			return m, fetchArticleCmd(m.Services, story)
		}

		return m, nil
	}
	if m.OverlayOpen {
		switch msg.String() {
		case "esc", "q":
			m.OverlayOpen = false
			m.OverlayText = ""
			m.LeaderPending = false
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil
		}
	}
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "h", "left":
		m.LeaderPending = false
		m.previousTopic()
		return m, loadTopicStoriesCmd(m.Services, m.CurrentTopic(), m.TimeRange)

	case "l", "right":
		m.LeaderPending = false
		m.nextTopic()
		return m, loadTopicStoriesCmd(m.Services, m.CurrentTopic(), m.TimeRange)

	case "j", "down":
		m.LeaderPending = false
		m.nextStory()

	case "k", "up":
		m.LeaderPending = false
		m.previousStory()

	case "t":
		m.LeaderPending = false
		m.TimeRange = m.TimeRange.Next()
		m.ActiveStory = 0
		m.StoryOffset = 0
		return m, loadTopicStoriesCmd(m.Services, m.CurrentTopic(), m.TimeRange)

	case ",":
		m.LeaderPending = true

	case " ":
		if m.LeaderPending {
			m.LeaderPending = false
			text, err := m.BuildSummary()
			if err != nil {
				m.Err = err
				return m, nil
			}

			m.OverlayText = text
			m.OverlayOpen = true
		}

	case "enter":
		m = m.OpenReader()
		return m, nil

	case "r":
		m.LeaderPending = false
		m.LoadingCount = 1
		return m, refreshTopicCmd(m.Services, m.CurrentTopic(), m.TimeRange)

	case "R":
		m.LeaderPending = false
		m.LoadingCount = len(m.Topics)
		return m, refreshAllTopicsCmd(m.Services, m.Topics, m.TimeRange)

	case "pgdown", "ctrl+d":
		m.LeaderPending = false
		m.pageDown()

	case "pgup", "ctrl+u":
		m.LeaderPending = false
		m.pageUp()

	case "f":
		story, ok := m.CurrentStory()
		if !ok {
			return m, nil
		}
		m.FetchingArticle = true
		m.Err = nil
		return m, fetchArticleCmd(m.Services, story)

	default:
		m.LeaderPending = false
	}

	return m, nil
}

func (m *Model) previousTopic() {
	m.ActiveStory = 0
	m.StoryOffset = 0
	if len(m.Topics) == 0 {
		return
	}

	if m.ActiveTopic > 0 {
		m.ActiveTopic--
	} else {
		m.ActiveTopic = len(m.Topics) - 1
	}

	m.ActiveStory = 0
	m.TimeRange = initialTimeRange(m.CurrentTopic())
}

func (m *Model) nextTopic() {
	m.ActiveStory = 0
	m.StoryOffset = 0
	if len(m.Topics) == 0 {
		return
	}

	if m.ActiveTopic < len(m.Topics)-1 {
		m.ActiveTopic++
	} else {
		m.ActiveTopic = 0
	}

	m.ActiveStory = 0
	m.TimeRange = initialTimeRange(m.CurrentTopic())
}

func (m *Model) previousStory() {
	if m.ActiveStory > 0 {
		m.ActiveStory--
	}

	m.ensureSelectedStoryVisible()
}

func (m *Model) nextStory() {
	stories := m.CurrentStories()
	if len(stories) == 0 {
		m.ActiveStory = 0
		m.StoryOffset = 0
		return
	}

	if m.ActiveStory < len(stories)-1 {
		m.ActiveStory++
	}

	m.ensureSelectedStoryVisible()
}

func (m *Model) ensureSelectedStoryVisible() {
	visible := m.visibleStoryCount()
	if visible <= 0 {
		return
	}

	if m.ActiveStory < m.StoryOffset {
		m.StoryOffset = m.ActiveStory
		return
	}

	if m.ActiveStory >= m.StoryOffset+visible {
		m.StoryOffset = m.ActiveStory - visible + 1
	}

	if m.StoryOffset < 0 {
		m.StoryOffset = 0
	}
}

func (m Model) visibleStoryCount() int {
	// Rough terminal budgeting:
	// header + tabs + spacing + "Stories" label + preview box + footer.
	// Tune this later.
	count := m.Height - 14
	if count < 5 {
		return 5
	}

	if count > 20 {
		return 20
	}

	return count
}

func (m *Model) pageDown() {
	stories := m.CurrentStories()
	if len(stories) == 0 {
		m.ActiveStory = 0
		m.StoryOffset = 0
		return
	}

	step := m.visibleStoryCount()
	if step < 1 {
		step = 5
	}

	m.ActiveStory += step
	if m.ActiveStory >= len(stories) {
		m.ActiveStory = len(stories) - 1
	}

	m.ensureSelectedStoryVisible()
}

func (m *Model) pageUp() {
	step := m.visibleStoryCount()
	if step < 1 {
		step = 5
	}

	m.ActiveStory -= step
	if m.ActiveStory < 0 {
		m.ActiveStory = 0
	}

	m.ensureSelectedStoryVisible()
}

func (m *Model) updateStory(updated core.Story) {
	stories := m.Stories[updated.TopicID]
	for i := range stories {
		if stories[i].ID == updated.ID {
			stories[i] = updated
			m.Stories[updated.TopicID] = stories
			return
		}
	}
}
