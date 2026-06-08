package app

import "github.com/patjrobinson/foxden/internal/core"

type topicStoriesLoadedMsg struct {
	TopicID string
	Stories []core.Story
}

type topicRefreshFailedMsg struct {
	TopicID string
	Err     error
}
