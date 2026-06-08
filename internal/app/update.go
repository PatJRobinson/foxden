package app

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "h", "left":
		m.LeaderPending = false
		m.previousTopic()

	case "l", "right":
		m.LeaderPending = false
		m.nextTopic()

	case "j", "down":
		m.LeaderPending = false
		m.nextStory()

	case "k", "up":
		m.LeaderPending = false
		m.previousStory()

	case "t":
		m.LeaderPending = false
		m.TimeRange = m.TimeRange.Next()

	case ",":
		m.LeaderPending = true

	case " ":
		if m.LeaderPending {
			m.LeaderPending = false
			// Commit 5 will open a summary overlay here.
			m.Err = nil
		}
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
