package core

type Topic struct {
	ID      string
	Title   string
	Sources []Source
	Summary SummaryConfig
	Agents  string
}
