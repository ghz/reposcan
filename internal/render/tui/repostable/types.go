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

	tbl table.Model

	displayMode tableDisplayMode

	// repos mode
	report        report.ScanReport
	filteredRepos []report.RepoState
	filterQuery   string
	favorites     map[string]bool // keyed by repo path

	// folders mode
	folders          []report.FolderEntry
	filteredFolders  []report.FolderEntry
	repoStatesByPath map[string]report.RepoState
}
