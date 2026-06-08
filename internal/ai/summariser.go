package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/patjrobinson/foxden/internal/core"
)

type SummaryRequest struct {
	TopicTitle string
	TimeRange  core.TimeRange
	AgentsMD   string
	Stories    []core.Story
}

type Summary struct {
	Text string
}

type Summariser interface {
	Summarise(ctx context.Context, req SummaryRequest) (Summary, error)
}

type PlaceholderSummariser struct{}

func (s PlaceholderSummariser) Summarise(ctx context.Context, req SummaryRequest) (Summary, error) {
	select {
	case <-ctx.Done():
		return Summary{}, ctx.Err()
	default:
	}

	agentsLoaded := "no"
	if strings.TrimSpace(req.AgentsMD) != "" {
		agentsLoaded = "yes"
	}

	text := fmt.Sprintf(
		"AI summary: %s / %s\n\nPlaceholder summary for %s.\nStories considered: %d.\nAGENTS.md loaded: %s.\n\nThis will eventually call a real summariser using the active topic, selected time range, story list, and topic-specific AGENTS.md instructions.\n\n[esc/q close]",
		req.TopicTitle,
		req.TimeRange,
		req.TopicTitle,
		len(req.Stories),
		agentsLoaded,
	)

	return Summary{Text: text}, nil
}
