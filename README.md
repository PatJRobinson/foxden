# news-tui

`news-tui` is an experimental terminal news reader and scraper frontend written in Go.

The aim is to build a local-first TUI news cockpit with configurable topic tabs. Each topic can pull from different kinds of sources: RSS feeds, web pages, GitHub releases, forums, working group notes, and eventually custom scrapers.

The app is intended to feel Vim-like:

* `h` / `l` switch between topic tabs
* `j` / `k` move through stories
* `t` cycles the active time range
* leader key + space opens an AI summary overlay
* `q` quits or closes an overlay

The long-term idea is that each topic can also include an optional `AGENTS.md` file. This file gives topic-specific instructions to the AI summarizer, such as what to prioritize, what to down-rank, and how to interpret news for that domain.

## Current status

This project is currently a working skeleton.

Implemented:

* Go module and Nix development shell
* Bubble Tea TUI shell
* Topic tabs loaded from YAML config
* Vim-style navigation
* Time range cycling
* Placeholder AI summary overlay
* Topic-level `AGENTS.md` loading
* Core domain models for topics, sources, stories, and time ranges
* Ingestion interfaces for future RSS/scraping support
* SQLite schema and story upsert path
* Basic tests for config loading, time range cycling, ingestion manager behavior, and storage

Not implemented yet:

* Real RSS fetching in the running TUI
* SQLite-backed story loading in the TUI
* Real scraping
* Real AI provider integration
* Opening stories in a browser
* Search
* Bookmarks
* Read/unread state

## Development

This project uses Go modules for Go dependencies and Nix for the development environment.

Enter the dev shell:

```bash
nix develop
```

Run the app:

```bash
go run ./cmd/news-tui
```

Or, if `just` is available:

```bash
just run
```

Run tests:

```bash
go test ./...
```

Or:

```bash
just test
```

Format and tidy:

```bash
gofmt -w .
go mod tidy
```

Or:

```bash
just check
```

## Nix

The project includes a `flake.nix` for a reproducible development shell.

Useful checks:

```bash
nix flake show
nix develop -c go test ./...
```

The flake may also define a `buildGoModule` package. If `vendorHash` is still set to `pkgs.lib.fakeHash`, the first `nix build` is expected to fail and print the correct hash. Replace the fake hash with the printed value, then run:

```bash
nix build
```

For pure-Go SQLite, the project currently prefers `modernc.org/sqlite` to avoid CGO complexity.

## Keybindings

| Key            | Action                              |
| -------------- | ----------------------------------- |
| `h` / left     | Previous topic tab                  |
| `l` / right    | Next topic tab                      |
| `j` / down     | Next story                          |
| `k` / up       | Previous story                      |
| `t`            | Cycle time range                    |
| `,`            | Leader key                          |
| `,` then space | Open placeholder AI summary overlay |
| `esc`          | Close overlay                       |
| `q`            | Close overlay, or quit from main UI |
| `ctrl+c`       | Quit                                |

## Topic configuration

Topics live under `topics/`.

Example:

```text
topics/
  hacker-news/
    topic.yaml
    AGENTS.md

  ros2/
    topic.yaml
    AGENTS.md
```

A topic config looks like this:

```yaml
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
```

`AGENTS.md` is optional. When present, its contents are loaded into the topic and passed to the summary layer. The current summary implementation is only a placeholder, but the interface is already designed so a real AI provider can use the topic instructions later.

## Time ranges

Supported time ranges:

* `day`
* `week`
* `month`
* `quarter`
* `year`
* `5y`

The `t` key cycles through these in order.

## Architecture

The code is split into small internal packages:

```text
cmd/news-tui/
  main.go              executable entrypoint

internal/app/
  Bubble Tea model, update, view, styles, and demo data

internal/core/
  Topic, Source, Story, and TimeRange domain models

internal/config/
  YAML topic loader and AGENTS.md loading

internal/ai/
  Summarizer interface and placeholder summarizer

internal/ingest/
  Fetcher interface and refresh manager

internal/store/
  SQLite open/init code, migrations, and story upsert path

topics/
  Example topic definitions
```

The intended flow is:

```text
topic config
→ source fetchers
→ normalized stories
→ SQLite store
→ TUI story lists
→ AI summary overlay
```

The TUI should remain separate from ingestion, storage, and AI provider details.

## Planned next steps

Likely next branches:

### RSS ingestion and SQLite-backed stories

* Implement an RSS fetcher using `gofeed`
* Generate stable story IDs
* Store fetched stories in SQLite
* Load stories from SQLite into the TUI
* Add `r` to refresh the active topic
* Add `R` to refresh all topics

### Story opening

* Add `enter` to open the selected story URL in the system browser
* Support Linux/macOS/BSD-friendly opener detection

### Real summary provider

* Add an OpenAI-compatible summarizer implementation
* Cache summaries by topic, time range, newest story timestamp, model, and AGENTS.md hash
* Keep placeholder/local summary as a fallback

### Better UI

* Improve layout sizing
* Make the summary modal feel like a real overlay
* Add a source/status line
* Add loading and error states
* Add search/filter mode

## Design notes

Prefer this source order:

```text
official API > RSS/Atom > structured HTML > custom scraper > browser automation
```

The app should remain local-first and inspectable. Scraping should be source-specific where possible, especially for complex topics like ROS2 where the meaning of a source matters as much as the raw text.

For now, the priority is to keep the bones clean:

* topic config is external
* ingestion is pluggable
* storage is isolated
* AI summarization is behind an interface
* the TUI is just a frontend over app state

