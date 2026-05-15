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

	repoState *report.RepoState
	theme     theme.Theme
}
