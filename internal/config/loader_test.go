package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTopics(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "ros2", "topic.yaml"), `
id: ros2
title: ROS2

sources:
  - id: ros-discourse
    name: ROS Discourse
    type: rss
    url: https://discourse.ros.org/c/ros2.rss

  - id: nav2-github
    name: Nav2 GitHub Releases
    type: github_releases
    repo: ros-navigation/navigation2

summary:
  agents: AGENTS.md
  default_range: week
`)

	writeFile(t, filepath.Join(root, "ros2", "AGENTS.md"), `
# ROS2 instructions

Prioritize working group decisions and migration notes.
`)

	writeFile(t, filepath.Join(root, "hacker-news", "topic.yaml"), `
id: hacker-news
title: Hacker News

sources:
  - id: hn-rss
    name: Hacker News RSS
    type: rss
    url: https://news.ycombinator.com/rss

summary:
  agents: AGENTS.md
  default_range: day
`)

	// This directory should be ignored because it has no topic.yaml.
	if err := os.MkdirAll(filepath.Join(root, "not-a-topic"), 0o755); err != nil {
		t.Fatalf("create ignored directory: %v", err)
	}

	topics, err := LoadTopics(root)
	if err != nil {
		t.Fatalf("LoadTopics() error = %v", err)
	}

	if len(topics) != 2 {
		t.Fatalf("LoadTopics() returned %d topics, want 2", len(topics))
	}

	// LoadTopics sorts by topic ID, so hacker-news should come first.
	hn := topics[0]
	if hn.ID != "hacker-news" {
		t.Fatalf("topics[0].ID = %q, want %q", hn.ID, "hacker-news")
	}

	if hn.Title != "Hacker News" {
		t.Fatalf("hacker-news title = %q, want %q", hn.Title, "Hacker News")
	}

	if hn.Summary.Agents != "AGENTS.md" {
		t.Fatalf("hacker-news summary agents = %q, want %q", hn.Summary.Agents, "AGENTS.md")
	}

	if hn.Summary.DefaultRange != "day" {
		t.Fatalf("hacker-news default range = %q, want %q", hn.Summary.DefaultRange, "day")
	}

	if len(hn.Sources) != 1 {
		t.Fatalf("hacker-news source count = %d, want 1", len(hn.Sources))
	}

	if hn.Sources[0].ID != "hn-rss" {
		t.Fatalf("hacker-news source ID = %q, want %q", hn.Sources[0].ID, "hn-rss")
	}

	if hn.Sources[0].URL != "https://news.ycombinator.com/rss" {
		t.Fatalf("hacker-news source URL = %q", hn.Sources[0].URL)
	}

	// Missing AGENTS.md should be tolerated, so Hacker News has no loaded
	// Agents body even though the config references AGENTS.md.
	if hn.Agents != "" {
		t.Fatalf("hacker-news Agents = %q, want empty string", hn.Agents)
	}

	ros2 := topics[1]
	if ros2.ID != "ros2" {
		t.Fatalf("topics[1].ID = %q, want %q", ros2.ID, "ros2")
	}

	if ros2.Title != "ROS2" {
		t.Fatalf("ros2 title = %q, want %q", ros2.Title, "ROS2")
	}

	if len(ros2.Sources) != 2 {
		t.Fatalf("ros2 source count = %d, want 2", len(ros2.Sources))
	}

	if ros2.Sources[0].ID != "ros-discourse" {
		t.Fatalf("ros2 first source ID = %q, want %q", ros2.Sources[0].ID, "ros-discourse")
	}

	if ros2.Sources[1].Repo != "ros-navigation/navigation2" {
		t.Fatalf("ros2 second source repo = %q, want %q", ros2.Sources[1].Repo, "ros-navigation/navigation2")
	}

	if !strings.Contains(ros2.Agents, "Prioritize working group decisions") {
		t.Fatalf("ros2 Agents did not contain expected instructions: %q", ros2.Agents)
	}
}

func TestLoadTopicsMissingRoot(t *testing.T) {
	_, err := LoadTopics(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("LoadTopics() error = nil, want error for missing root")
	}
}

func TestLoadTopicsRejectsMissingRequiredTopicFields(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "broken", "topic.yaml"), `
title: Broken Topic
`)

	_, err := LoadTopics(root)
	if err == nil {
		t.Fatal("LoadTopics() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "missing required field id") {
		t.Fatalf("LoadTopics() error = %q, want missing id error", err)
	}
}

func TestLoadTopicsRejectsInvalidSource(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "broken", "topic.yaml"), `
id: broken
title: Broken Topic

sources:
  - id: bad-source
    name: Bad Source
`)

	_, err := LoadTopics(root)
	if err == nil {
		t.Fatal("LoadTopics() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "missing required field type") {
		t.Fatalf("LoadTopics() error = %q, want missing source type error", err)
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %q: %v", path, err)
	}

	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
