package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type DetailsSubMode int

const (
	DetailsSubModeFiles   DetailsSubMode = 0
	DetailsSubModeDiff    DetailsSubMode = 1
	DetailsSubModeCommits DetailsSubMode = 2
	DetailsSubModeReadme  DetailsSubMode = 3

	detailsSubModeCount = 4
)

type Model struct {
	width   int
	height  int
	subMode DetailsSubMode
	commits []string
	readme  []string
	diff    []string

	// scrollOffset is the index of the first body line shown in the panel.
	// It is reset to 0 whenever the panel's content changes (tab switch or a
	// different repo selected).
	scrollOffset int

	repoState *report.RepoState
	theme     theme.Theme
}
