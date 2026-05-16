package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type DetailsSubMode int

const (
	DetailsSubModeFiles   DetailsSubMode = 0
	DetailsSubModeCommits DetailsSubMode = 1
	DetailsSubModeReadme  DetailsSubMode = 2

	detailsSubModeCount = 3
)

type Model struct {
	width   int
	height  int
	subMode DetailsSubMode
	commits []string
	readme  []string

	// scrollOffset is the index of the first body line shown in the panel.
	// It is reset to 0 whenever the panel's content changes (tab switch or a
	// different repo selected).
	scrollOffset int

	repoState *report.RepoState
	theme     theme.Theme
}
