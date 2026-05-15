package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type DetailsSubMode int

const (
	DetailsSubModeFiles   DetailsSubMode = 0
	DetailsSubModeCommits DetailsSubMode = 1
)

type Model struct {
	height  int
	subMode DetailsSubMode
	commits []string

	repoState *report.RepoState
	theme     theme.Theme
}
