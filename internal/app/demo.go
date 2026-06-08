package app

import (
	"fmt"
	"time"

	"github.com/patjrobinson/news-tui/internal/core"
)

func DemoTopics() []core.Topic {
	return []core.Topic{
		{
			ID:    "hacker-news",
			Title: "Hacker News",
			Summary: core.SummaryConfig{
				Agents:       "AGENTS.md",
				DefaultRange: "day",
			},
		},
		{
			ID:    "ros2",
			Title: "ROS2",
			Summary: core.SummaryConfig{
				Agents:       "AGENTS.md",
				DefaultRange: "week",
			},
		},
	}
}

func DemoStories(topics []core.Topic) map[string][]core.Story {
	now := time.Now()
	stories := make(map[string][]core.Story, len(topics))

	for _, topic := range topics {
		stories[topic.ID] = []core.Story{
			{
				ID:          fmt.Sprintf("%s-1", topic.ID),
				TopicID:     topic.ID,
				SourceID:    "demo",
				SourceName:  "Demo Source",
				Title:       fmt.Sprintf("%s: first demo story", topic.Title),
				URL:         "https://example.com/first",
				PublishedAt: now.Add(-2 * time.Hour),
				FetchedAt:   now,
				Excerpt:     "This is placeholder content showing how the preview pane will feel.",
			},
			{
				ID:          fmt.Sprintf("%s-2", topic.ID),
				TopicID:     topic.ID,
				SourceID:    "demo",
				SourceName:  "Demo Source",
				Title:       fmt.Sprintf("%s: second demo story", topic.Title),
				URL:         "https://example.com/second",
				PublishedAt: now.Add(-8 * time.Hour),
				FetchedAt:   now,
				Excerpt:     "A second story gives us something to move to with j/k navigation.",
			},
			{
				ID:          fmt.Sprintf("%s-3", topic.ID),
				TopicID:     topic.ID,
				SourceID:    "demo",
				SourceName:  "Demo Source",
				Title:       fmt.Sprintf("%s: third demo story", topic.Title),
				URL:         "https://example.com/third",
				PublishedAt: now.Add(-24 * time.Hour),
				FetchedAt:   now,
				Excerpt:     "Later, these stories will come from RSS, scraping, and persisted SQLite rows.",
			},
		}
	}

	return stories
}
