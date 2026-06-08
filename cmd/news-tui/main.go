package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patjrobinson/news-tui/internal/app"
	"github.com/patjrobinson/news-tui/internal/config"
	"github.com/patjrobinson/news-tui/internal/store"
)

func main() {
	ctx := context.Background()

	topics, err := config.LoadTopics("topics")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load topics: %v\n", err)
		topics = app.DemoTopics()
	}

	db, err := store.Open(".news-tui.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to open database: %v\n", err)
	} else {
		defer db.Close()

		if err := db.Init(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to initialize database: %v\n", err)
			_ = db.Close()
			db = nil
		}
	}

	services := app.NewServices(db)
	model := app.NewModel(topics, services)

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running news-tui: %v\n", err)
		os.Exit(1)
	}
}
