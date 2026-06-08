package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patjrobinson/news-tui/internal/app"
	"github.com/patjrobinson/news-tui/internal/config"
)

func main() {
	topics, err := config.LoadTopics("topics")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load topics: %v\n", err)
		topics = app.DemoTopics()
	}

	model := app.NewModel(topics)

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running news-tui: %v\n", err)
		os.Exit(1)
	}
}
