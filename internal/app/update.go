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
		m.Stories[msg.TopicID] = msg.Stories
		m.Err = nil

		if m.ActiveStory >= len(m.CurrentStories()) {
			m.ActiveStory = 0
		}

		return m, nil

	case topicRefreshFailedMsg:
		m.Err = msg.Err
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case "r":
		m.LeaderPending = false
		return m, refreshTopicCmd(m.Services, m.CurrentTopic(), m.TimeRange)

	case "R":
		m.LeaderPending = false
		return m, refreshAllTopicsCmd(m.Services, m.Topics, m.TimeRange)

	default:
		m.LeaderPending = false
	}

	return m, nil
}

func (m *Model) previousTopic() {
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
}

func (m *Model) nextStory() {
	stories := m.CurrentStories()
	if len(stories) == 0 {
		m.ActiveStory = 0
		return
	}

	if m.ActiveStory < len(stories)-1 {
		m.ActiveStory++
	}
}
