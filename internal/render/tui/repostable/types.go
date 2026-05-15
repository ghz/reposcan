package repostable

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type tableDisplayMode int

const (
	tableDisplayRepos   tableDisplayMode = 0
	tableDisplayFolders tableDisplayMode = 1
)

// rowRef maps a visual table row back to its source. In repos mode the table
// is a tree: each repo header row may be followed by indented branch rows.
type rowRef struct {
	repoIdx  int // index into filteredRepos
	branchIx int // -1 for the repo header row; otherwise index into branchCache[repoID]
}

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

	expanded    map[string]bool                // repo ID -> branch rows shown
	branchCache map[string][]gitx.BranchStatus // repo ID -> lazily fetched branches
	rowRefs     []rowRef                       // parallel to table rows (repos mode)

	// folders mode
	folders          []report.FolderEntry
	filteredFolders  []report.FolderEntry
	repoStatesByPath map[string]report.RepoState
}
