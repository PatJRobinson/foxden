# foxden

`foxden` is an experimental terminal news reader and scraper frontend written in Go.

The aim is to build a local-first TUI news cockpit with configurable topic tabs. Each topic can pull from different kinds of sources: RSS feeds, web pages, GitHub releases, forums, working group notes, and eventually custom scrapers.

The app is for keeping up with the outside world from a safe little terminal den: fewer websites, less context switching, more low-effort signal.

The app is intended to feel Vim-like:

* `h` / `l` switch between topic tabs
* `j` / `k` move through stories
* `enter` opens the selected story in the in-app reader
* `t` cycles the active time range
* `r` refreshes the active topic
* `R` refreshes all topics
* leader key + space opens an AI summary overlay
* `q` quits, closes an overlay, or exits the reader

The long-term idea is that each topic can also include an optional `AGENTS.md` file. This file gives topic-specific instructions to the AI summarizer, such as what to prioritize, what to down-rank, and how to interpret news for that domain.

## Current status

This project is currently a working early prototype.

Implemented:

* Go module and Nix development shell
* Bubble Tea TUI shell
* Topic tabs loaded from YAML config
* Vim-style navigation
* Scrolling story list for long feeds
* In-app reader for stored story content
* Time range cycling
* RSS feed ingestion
* GitHub releases ingestion
* SQLite-backed story persistence
* Refresh commands for active topic and all topics
* Placeholder AI summary overlay
* Topic-level `AGENTS.md` loading
* Core domain models for topics, sources, stories, and time ranges
* Pluggable ingestion interfaces
* SQLite schema and story upsert/query paths
* Basic tests for config loading, time range cycling, ingestion manager behavior, storage, and reader wrapping

Not implemented yet:

* Real web scraping
* Original article fetching/extraction for headline-only feeds
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
go run ./cmd/foxden
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

### Main story list

| Key                  | Action                                            |
| -------------------- | ------------------------------------------------- |
| `h` / left           | Previous topic tab                                |
| `l` / right          | Next topic tab                                    |
| `j` / down           | Next story                                        |
| `k` / up             | Previous story                                    |
| `ctrl+d` / page down | Move down by a page in the story list, if enabled |
| `ctrl+u` / page up   | Move up by a page in the story list, if enabled   |
| `t`                  | Cycle time range                                  |
| `r`                  | Refresh active topic                              |
| `R`                  | Refresh all topics                                |
| `enter`              | Open selected story in the reader                 |
| `,`                  | Leader key                                        |
| `,` then space       | Open placeholder AI summary overlay               |
| `q`                  | Quit from main UI                                 |
| `ctrl+c`             | Quit                                              |

### Reader

| Key                  | Action         |
| -------------------- | -------------- |
| `j` / down           | Scroll down    |
| `k` / up             | Scroll up      |
| `ctrl+d` / page down | Page down      |
| `ctrl+u` / page up   | Page up        |
| `g`                  | Jump to top    |
| `G`                  | Jump to bottom |
| `esc`                | Close reader   |
| `q`                  | Close reader   |
| `ctrl+c`             | Quit           |

### Summary overlay

| Key      | Action        |
| -------- | ------------- |
| `esc`    | Close overlay |
| `q`      | Close overlay |
| `ctrl+c` | Quit          |

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
    name: Open Robotics Discourse
    type: rss
    url: https://discourse.openrobotics.org/latest.rss

  - id: nav2-github
    name: Nav2 GitHub Releases
    type: github_releases
    repo: ros-navigation/navigation2

summary:
  agents: AGENTS.md
  default_range: week
```

Supported source types currently include:

| Type              | Meaning                                                                 |
| ----------------- | ----------------------------------------------------------------------- |
| `rss`             | Fetch RSS/Atom feed items and normalize them into stories               |
| `github_releases` | Fetch releases from a GitHub repository and normalize them into stories |

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

Time ranges are used when loading stories from SQLite. They do not yet control how much remote feed history is fetched from each source.

## In-app reader

Press `enter` on a selected story to open it in the reader.

The reader displays the stored story content from the local database. For sources like Discourse RSS or GitHub releases, this can include useful body text. For headline-only feeds such as Hacker News RSS, the stored content may only be a comments marker or short feed description until original article fetching is implemented.

This means the current reader is best understood as:

```text
stored feed/release content reader
```

not yet:

```text
full article readability extractor
```

## RSS and GitHub release ingestion

The app can refresh configured sources and store normalized stories locally.

Refresh commands:

* `r`: refresh active topic
* `R`: refresh all topics

RSS feed items are cleaned into plain text before storage. This is especially useful for HTML-heavy feeds such as Discourse, where RSS descriptions often contain markup.

GitHub release sources use the repository configured in `topic.yaml`, for example:

```yaml
- id: nav2-github
  name: Nav2 GitHub Releases
  type: github_releases
  repo: ros-navigation/navigation2
```

## SQLite-backed persistence

The app currently stores development data in:

```text
.foxden.db
```

in the project root.

This is intentionally temporary. Later, the database should move to an XDG data directory.

Stories are fetched from configured sources, normalized, upserted into SQLite, and then loaded back from SQLite into the TUI according to the active topic and time range.

## Architecture

The code is split into small internal packages:

```text
cmd/foxden/
  main.go              executable entrypoint

internal/app/
  Bubble Tea model, update, view, styles, reader, and UI commands

internal/core/
  Topic, Source, Story, and TimeRange domain models

internal/config/
  YAML topic loader and AGENTS.md loading

internal/ai/
  Summarizer interface and placeholder summarizer

internal/ingest/
  Fetcher interface, refresh manager, RSS fetcher, and GitHub releases fetcher

internal/store/
  SQLite open/init code, migrations, story upsert, and story query paths

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
→ in-app reader / AI summary overlay
```

The TUI should remain separate from ingestion, storage, and AI provider details.

## Planned next steps

Likely next branches:

### External story opening

* Add `o` to open the selected story URL in the system browser
* Support Linux/macOS/BSD-friendly opener detection
* Keep `enter` for in-app reader

### Original article fetching

* Fetch original article URLs for headline-only sources
* Extract readable article text
* Store extracted text in SQLite
* Show extracted content in the in-app reader

### Real summary provider

* Add an OpenAI-compatible summarizer implementation
* Cache summaries by topic, time range, newest story timestamp, model, and AGENTS.md hash
* Keep placeholder/local summary as a fallback

### Better UI

* Improve layout sizing
* Make the summary modal feel like a real overlay
* Add a source/status line
* Improve loading and error states
* Add search/filter mode

### Reading workflow

* Add read/unread state
* Add bookmarks
* Add story search
* Add per-topic sort/scoring options

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

