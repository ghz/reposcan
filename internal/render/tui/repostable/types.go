package repostable

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type tableDisplayMode int

const (
	tableDisplayRepos   tableDisplayMode = 0
	tableDisplayFolders tableDisplayMode = 1
)

type Model struct {
	width  int
	height int
	theme  theme.Theme

	tbl   table.Model
	title string

	displayMode tableDisplayMode

	// repos mode
	report        report.ScanReport
	filteredRepos []report.RepoState
	filterQuery   string

	// folders mode
	folders          []report.FolderEntry
	repoStatesByPath map[string]report.RepoState
}
