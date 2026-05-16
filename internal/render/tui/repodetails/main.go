package repodetails

import (
	"os"
	"path/filepath"
	"strings"

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

func (m *Model) UpdateSize(width, height int) {
	m.width = width
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

// CycleSubMode moves through the files, commits and readme views. forward
// advances to the next tab, otherwise it goes back to the previous one. rs is
// the currently selected repo and is used to load data for the new mode.
func (m *Model) CycleSubMode(rs *report.RepoState, forward bool) {
	if forward {
		m.subMode = (m.subMode + 1) % detailsSubModeCount
	} else {
		m.subMode = (m.subMode - 1 + detailsSubModeCount) % detailsSubModeCount
	}
	m.scrollOffset = 0
	m.loadForSubMode(rs)
}

// visibleRows is the number of body lines the details panel can show at once,
// excluding the "Path:" and tab-bar header lines and the scroll-hint line.
func (m *Model) visibleRows() int {
	rows := m.height - 3
	if rows < 1 {
		rows = 1
	}
	return rows
}

// contentLength returns the number of body lines for the active sub-mode.
func (m *Model) contentLength() int {
	switch m.subMode {
	case DetailsSubModeCommits:
		return len(m.commits)
	case DetailsSubModeReadme:
		return len(m.readme)
	default:
		if m.repoState == nil {
			return 0
		}
		return len(m.repoState.UncommitedFiles)
	}
}

// maxScroll is the largest valid scroll offset for the active sub-mode.
func (m *Model) maxScroll() int {
	max := m.contentLength() - m.visibleRows()
	if max < 0 {
		return 0
	}
	return max
}

// ScrollPageDown moves the details viewport down by one page.
func (m *Model) ScrollPageDown() {
	m.scrollOffset += m.visibleRows()
	if m.scrollOffset > m.maxScroll() {
		m.scrollOffset = m.maxScroll()
	}
}

// ScrollPageUp moves the details viewport up by one page.
func (m *Model) ScrollPageUp() {
	m.scrollOffset -= m.visibleRows()
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// ResetScroll returns the details viewport to the top. It is called whenever
// the panel's content changes (tab switch or a different repo selected).
func (m *Model) ResetScroll() {
	m.scrollOffset = 0
}

// ReloadForRepo reloads the data backing the current sub-mode (called when the
// cursor moves to a different repo).
func (m *Model) ReloadForRepo(rs *report.RepoState) {
	m.loadForSubMode(rs)
}

// loadForSubMode fetches the data required by the active sub-mode and clears
// data for the other modes so it stays in sync with the selected repo.
func (m *Model) loadForSubMode(rs *report.RepoState) {
	m.commits = nil
	m.readme = nil
	switch m.subMode {
	case DetailsSubModeCommits:
		m.fetchCommitsForRepo(rs)
	case DetailsSubModeReadme:
		m.fetchReadmeForRepo(rs)
	}
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

// readmeCandidates lists the file names looked up for a repo's README, in
// order of preference.
var readmeCandidates = []string{
	"README.md", "README.MD", "readme.md",
	"README", "README.txt", "README.rst",
}

func (m *Model) fetchReadmeForRepo(rs *report.RepoState) {
	if rs == nil {
		m.readme = nil
		return
	}
	for _, name := range readmeCandidates {
		data, err := os.ReadFile(filepath.Join(rs.Path, name))
		if err != nil {
			continue
		}
		normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
		m.readme = strings.Split(strings.TrimRight(normalized, "\n"), "\n")
		return
	}
	m.readme = []string{}
}

func (m Model) Init() tea.Cmd { return nil }
