package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/patjrobinson/news-tui/internal/core"
)

type Model struct {
	Topics []core.Topic

	ActiveTopic int
	ActiveStory int
	TimeRange   core.TimeRange

	Stories map[string][]core.Story

	Width  int
	Height int

	LeaderPending bool

	Err error
}

func NewModel(topics []core.Topic) Model {
	if len(topics) == 0 {
		topics = DemoTopics()
	}

	stories := DemoStories(topics)

	return Model{
		Topics:      topics,
		ActiveTopic: 0,
		ActiveStory: 0,
		TimeRange:   initialTimeRange(topics[0]),
		Stories:     stories,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) CurrentTopic() core.Topic {
	if len(m.Topics) == 0 {
		return core.Topic{}
	}

	if m.ActiveTopic < 0 || m.ActiveTopic >= len(m.Topics) {
		return m.Topics[0]
	}

	return m.Topics[m.ActiveTopic]
}

func (m Model) CurrentStories() []core.Story {
	topic := m.CurrentTopic()
	return m.Stories[topic.ID]
}

func (m Model) CurrentStory() (core.Story, bool) {
	stories := m.CurrentStories()
	if len(stories) == 0 {
		return core.Story{}, false
	}

	if m.ActiveStory < 0 || m.ActiveStory >= len(stories) {
		return stories[0], true
	}

	return stories[m.ActiveStory], true
}

func initialTimeRange(topic core.Topic) core.TimeRange {
	switch topic.Summary.DefaultRange {
	case string(core.TimeRangeDay):
		return core.TimeRangeDay
	case string(core.TimeRangeWeek):
		return core.TimeRangeWeek
	case string(core.TimeRangeMonth):
		return core.TimeRangeMonth
	case string(core.TimeRangeQuarter):
		return core.TimeRangeQuarter
	case string(core.TimeRangeYear):
		return core.TimeRangeYear
	case string(core.TimeRangeFiveY):
		return core.TimeRangeFiveY
	default:
		return core.TimeRangeWeek
	}
}
