package repodetails

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	repoState *report.RepoState,
	theme theme.Theme,
) Model {
	return Model{
		theme:     theme,
		repoState: repoState,
	}
}

func (m *Model) UpdateSize(height int) {
	m.height = height
}

// UpdateData is called from View() to keep the displayed path current.
// It only updates the repoState pointer — commit fetching is handled in the
// Update cycle via ToggleSubMode / RefetchCommits.
func (m *Model) UpdateData(repoState *report.RepoState) {
	m.repoState = repoState
}

func (m *Model) SubMode() DetailsSubMode {
	return m.subMode
}

// ToggleSubMode switches between files and commits view. rs is the currently
// selected repo and is used to fetch commits when switching to commits mode.
func (m *Model) ToggleSubMode(rs *report.RepoState) {
	if m.subMode == DetailsSubModeFiles {
		m.subMode = DetailsSubModeCommits
		m.fetchCommitsForRepo(rs)
	} else {
		m.subMode = DetailsSubModeFiles
		m.commits = nil
	}
}

// RefetchCommits re-fetches commits for the given repo (called when the cursor
// moves while in commits mode).
func (m *Model) RefetchCommits(rs *report.RepoState) {
	m.fetchCommitsForRepo(rs)
}

func (m *Model) fetchCommitsForRepo(rs *report.RepoState) {
	if rs == nil {
		m.commits = nil
		return
	}
	commits, err := gitx.GetRecentCommits(rs.Path, 30)
	if err != nil {
		m.commits = []string{}
		return
	}
	m.commits = commits
}

func (m Model) Init() tea.Cmd { return nil }
