package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patjrobinson/news-tui/internal/ai"
	"github.com/patjrobinson/news-tui/internal/core"
)

type Model struct {
	Topics []core.Topic

	ActiveTopic int
	ActiveStory int
	StoryOffset int
	TimeRange   core.TimeRange

	Stories map[string][]core.Story

	Width  int
	Height int

	LeaderPending bool

	OverlayOpen bool
	OverlayText string
	Summariser  ai.Summariser

	Services     Services
	LoadingCount int

	Err error
}

func (m Model) IsLoading() bool {
	return m.LoadingCount > 0
}

func NewModel(topics []core.Topic, servicesArg ...Services) Model {
	if len(topics) == 0 {
		topics = DemoTopics()
	}

	var services Services
	if len(servicesArg) > 0 {
		services = servicesArg[0]
	}

	stories := DemoStories(topics)

	loadingCount := 0
	if services.Store != nil {
		loadingCount = len(topics)
	}

	return Model{
		Topics:       topics,
		ActiveTopic:  0,
		ActiveStory:  0,
		TimeRange:    initialTimeRange(topics[0]),
		Stories:      stories,
		Summariser:   ai.PlaceholderSummariser{},
		Services:     services,
		LoadingCount: loadingCount,
	}
}

func (m Model) Init() tea.Cmd {
	if m.Services.Store == nil {
		return nil
	}

	return refreshAllTopicsCmd(m.Services, m.Topics, m.TimeRange)
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

func (m Model) BuildSummary() (string, error) {
	topic := m.CurrentTopic()

	summary, err := m.Summariser.Summarise(context.Background(), ai.SummaryRequest{
		TopicTitle: topic.Title,
		TimeRange:  m.TimeRange,
		AgentsMD:   topic.Agents,
		Stories:    m.CurrentStories(),
	})
	if err != nil {
		return "", err
	}

	return summary.Text, nil
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
