package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patjrobinson/foxden/internal/core"
)

func refreshTopicCmd(services Services, topic core.Topic, r core.TimeRange) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if _, err := services.RefreshTopic(ctx, topic); err != nil {
			return topicRefreshFailedMsg{
				TopicID: topic.ID,
				Err:     err,
			}
		}

		stories, err := services.LoadStories(ctx, topic, r)
		if err != nil {
			return topicRefreshFailedMsg{
				TopicID: topic.ID,
				Err:     err,
			}
		}

		return topicStoriesLoadedMsg{
			TopicID: topic.ID,
			Stories: stories,
		}
	}
}

func refreshAllTopicsCmd(services Services, topics []core.Topic, r core.TimeRange) tea.Cmd {
	var cmds []tea.Cmd

	for _, topic := range topics {
		topic := topic
		cmds = append(cmds, refreshTopicCmd(services, topic, r))
	}

	return tea.Batch(cmds...)
}

func loadTopicStoriesCmd(services Services, topic core.Topic, r core.TimeRange) tea.Cmd {
	return func() tea.Msg {
		stories, err := services.LoadStories(context.Background(), topic, r)
		if err != nil {
			return topicRefreshFailedMsg{
				TopicID: topic.ID,
				Err:     err,
			}
		}

		return topicStoriesLoadedMsg{
			TopicID: topic.ID,
			Stories: stories,
		}
	}
}
