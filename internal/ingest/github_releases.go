package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/patjrobinson/foxden/internal/core"
)

type GitHubReleasesFetcher struct {
	Client *http.Client
}

type githubRelease struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	TagName     string     `json:"tag_name"`
	HTMLURL     string     `json:"html_url"`
	Body        string     `json:"body"`
	Draft       bool       `json:"draft"`
	Prerelease  bool       `json:"prerelease"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   *time.Time `json:"created_at"`
	Author      *struct {
		Login string `json:"login"`
	} `json:"author"`
}

func NewGitHubReleasesFetcher() GitHubReleasesFetcher {
	return GitHubReleasesFetcher{
		Client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (f GitHubReleasesFetcher) Fetch(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error) {
	repo := strings.TrimSpace(source.Repo)
	if repo == "" {
		return nil, fmt.Errorf("github releases source %q is missing repo", source.ID)
	}

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=30", repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create github releases request for %q: %w", repo, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "foxden/0.1")

	// Optional: allows higher rate limits if the user exports GITHUB_TOKEN.
	if token := strings.TrimSpace(getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch github releases for %q: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch github releases for %q: unexpected status %s", repo, resp.Status)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode github releases for %q: %w", repo, err)
	}

	now := time.Now()
	stories := make([]core.Story, 0, len(releases))

	for _, release := range releases {
		if release.Draft {
			continue
		}

		releaseURL := strings.TrimSpace(release.HTMLURL)
		if releaseURL == "" {
			continue
		}

		title := strings.TrimSpace(release.Name)
		if title == "" {
			title = strings.TrimSpace(release.TagName)
		}
		if title == "" {
			title = fmt.Sprintf("Release %d", release.ID)
		}

		if release.Prerelease {
			title = title + " (pre-release)"
		}

		publishedAt := time.Time{}
		if release.PublishedAt != nil {
			publishedAt = *release.PublishedAt
		} else if release.CreatedAt != nil {
			publishedAt = *release.CreatedAt
		}

		author := ""
		if release.Author != nil {
			author = release.Author.Login
		}

		body := cleanGitHubReleaseText(release.Body)

		stories = append(stories, core.Story{
			ID:          githubReleaseStoryID(topic.ID, source.ID, releaseURL),
			TopicID:     topic.ID,
			SourceID:    source.ID,
			SourceName:  source.Name,
			Title:       title,
			URL:         releaseURL,
			Author:      author,
			PublishedAt: publishedAt,
			FetchedAt:   now,
			Excerpt:     truncateText(body, 500),
			Content:     body,
			Tags:        []string{"github", "release", repo},
		})
	}

	return stories, nil
}

func cleanGitHubReleaseText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	// GitHub release bodies are Markdown, not HTML. This is intentionally light:
	// remove obvious markdown noise, unescape entities, and normalize whitespace.
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = html.UnescapeString(value)
	value = whitespaceRE.ReplaceAllString(value, " ")

	return strings.TrimSpace(value)
}

func githubReleaseStoryID(topicID string, sourceID string, url string) string {
	sum := sha256.Sum256([]byte(topicID + "\x00" + sourceID + "\x00" + url))
	return hex.EncodeToString(sum[:])
}
