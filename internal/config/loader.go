package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/patjrobinson/foxden/internal/core"
)

type topicFile struct {
	ID      string            `yaml:"id"`
	Title   string            `yaml:"title"`
	Sources []sourceFile      `yaml:"sources"`
	Summary summaryConfigFile `yaml:"summary"`
}

type sourceFile struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	URL  string `yaml:"url"`
	Repo string `yaml:"repo"`
}

type summaryConfigFile struct {
	Agents       string `yaml:"agents"`
	DefaultRange string `yaml:"default_range"`
}

// LoadTopics reads topic definitions from immediate child directories of root.
//
// Each topic directory is expected to contain a topic.yaml file. If the topic's
// summary config references an AGENTS.md file, the file is read relative to the
// topic directory. Missing AGENTS.md files are tolerated; malformed topic.yaml
// files are not.
func LoadTopics(root string) ([]core.Topic, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read topics root %q: %w", root, err)
	}

	var topics []core.Topic

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(root, entry.Name())
		topicPath := filepath.Join(dir, "topic.yaml")

		if _, err := os.Stat(topicPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("stat topic file %q: %w", topicPath, err)
		}

		topic, err := loadTopicDir(dir)
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].ID < topics[j].ID
	})

	return topics, nil
}

func loadTopicDir(dir string) (core.Topic, error) {
	topicPath := filepath.Join(dir, "topic.yaml")

	data, err := os.ReadFile(topicPath)
	if err != nil {
		return core.Topic{}, fmt.Errorf("read topic file %q: %w", topicPath, err)
	}

	var tf topicFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return core.Topic{}, fmt.Errorf("parse topic file %q: %w", topicPath, err)
	}

	if tf.ID == "" {
		return core.Topic{}, fmt.Errorf("topic file %q is missing required field id", topicPath)
	}

	if tf.Title == "" {
		return core.Topic{}, fmt.Errorf("topic file %q is missing required field title", topicPath)
	}

	topic := core.Topic{
		ID:    tf.ID,
		Title: tf.Title,
		Summary: core.SummaryConfig{
			Agents:       tf.Summary.Agents,
			DefaultRange: tf.Summary.DefaultRange,
		},
		Sources: make([]core.Source, 0, len(tf.Sources)),
	}

	for _, sf := range tf.Sources {
		if sf.ID == "" {
			return core.Topic{}, fmt.Errorf("topic file %q has a source missing required field id", topicPath)
		}

		if sf.Name == "" {
			return core.Topic{}, fmt.Errorf("topic file %q has source %q missing required field name", topicPath, sf.ID)
		}

		if sf.Type == "" {
			return core.Topic{}, fmt.Errorf("topic file %q has source %q missing required field type", topicPath, sf.ID)
		}

		topic.Sources = append(topic.Sources, core.Source{
			ID:   sf.ID,
			Name: sf.Name,
			Type: sf.Type,
			URL:  sf.URL,
			Repo: sf.Repo,
		})
	}

	if tf.Summary.Agents != "" {
		agentsPath := filepath.Join(dir, tf.Summary.Agents)

		agentsData, err := os.ReadFile(agentsPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return core.Topic{}, fmt.Errorf("read agents file %q: %w", agentsPath, err)
			}
		} else {
			topic.Agents = string(agentsData)
		}
	}

	return topic, nil
}
