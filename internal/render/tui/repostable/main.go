// Package repostable is a Model that renders git repo states in a table. Providing functionality like filterning
package repostable

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	theme theme.Theme,
	report report.ScanReport,
	width int,
	height int,
) Model {
	model := Model{
		width:         width,
		height:        height,
		theme:         theme,
		report:        report,
		filteredRepos: report.RepoStates,
		filterQuery:   "",
		displayMode:   tableDisplayRepos,
	}

	cols := createColumns(width)
	rows := createRows(model.report.RepoStates, theme)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(height),
	)
	t.Focus()

	km := table.DefaultKeyMap()
	setKeymaps(km)

	if len(rows) == 0 {
		t.SetRows([]table.Row{{"", "", ""}})
	}

	t.SetStyles(table.Styles{
		Header:   model.theme.Styles.TableHeader,
		Selected: model.theme.Styles.TableSelectedRow,
		Cell:     model.theme.Styles.TableRow,
	})
	model.tbl = t

	return model
}

func (rt Model) Init() tea.Cmd { return nil }

func (m *Model) SetReport(report report.ScanReport) {
	m.report = report
	m.displayMode = tableDisplayRepos
	m.Filter(m.filterQuery)
}

func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetFolders switches the table to folder-display mode, showing all direct
// subdirectories with a visual indicator of whether each is a Git repo.
func (m *Model) SetFolders(folders []report.FolderEntry, repoStates []report.RepoState) {
	m.folders = folders
	m.displayMode = tableDisplayFolders

	m.repoStatesByPath = make(map[string]report.RepoState, len(repoStates))
	for _, rs := range repoStates {
		m.repoStatesByPath[rs.Path] = rs
	}

	rows := createFolderRows(folders, m.repoStatesByPath, m.theme)
	if len(rows) == 0 {
		rows = []table.Row{{"", "", ""}}
	}
	m.tbl.SetRows(rows)
	m.tbl.SetCursor(0)
}

func (m *Model) UpdateWindowSize(width int, height int) Model {
	m.width = width - 2
	m.height = height - 2

	m.tbl.SetHeight(m.height)
	cols := createColumns(m.width)
	m.tbl.SetColumns(cols)

	return *m
}

// Filter filters repo states based on repo name. Then update table based on filtered repos
func (m *Model) Filter(query string) {
	m.filterQuery = query
	q := strings.ToLower(strings.TrimSpace(query))
	if len(q) == 0 {
		m.filteredRepos = m.report.RepoStates
	} else {
		m.filteredRepos = []report.RepoState{}
		for _, rs := range m.report.RepoStates {
			if strings.Contains(strings.ToLower(rs.Repo), q) ||
				strings.Contains(strings.ToLower(rs.Branch), q) {
				m.filteredRepos = append(m.filteredRepos, rs)
			}
		}
	}

	cursorPosition := m.tbl.Cursor()

	rows := createRows(m.filteredRepos, m.theme)
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredRepos) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}
}

func (m *Model) UpdateRepoState(index int, newState report.RepoState) {
	m.filteredRepos[index] = newState

	originalIndex := getRepoIndex(m.report.RepoStates, newState.ID)
	if originalIndex != -1 {
		m.report.RepoStates[originalIndex] = newState
	}

	rows := createRows(m.filteredRepos, m.theme)
	m.tbl.SetRows(rows)
}

func (m *Model) Blur() {
	m.tbl.Blur()
}

func (m *Model) Focus() {
	m.tbl.Focus()
}

func (m *Model) Cursor() int {
	return m.tbl.Cursor()
}

func (rt *Model) ReposCount() int {
	if rt.displayMode == tableDisplayFolders {
		return len(rt.folders)
	}
	return len(rt.filteredRepos)
}

func (m *Model) GetCurrentRepoState() *report.RepoState {
	if m.displayMode == tableDisplayFolders {
		return m.GetCurrentFolderRepoState()
	}
	return m.GetRepoStateAt(m.Cursor())
}

func (m *Model) GetRepoStateAt(index int) *report.RepoState {
	if index < 0 || index >= len(m.filteredRepos) {
		return nil
	}
	return &m.filteredRepos[index]
}

// GetCurrentFolderEntry returns the FolderEntry at the current cursor in folders mode.
func (m *Model) GetCurrentFolderEntry() *report.FolderEntry {
	if m.displayMode != tableDisplayFolders {
		return nil
	}
	i := m.tbl.Cursor()
	if i < 0 || i >= len(m.folders) {
		return nil
	}
	return &m.folders[i]
}

// GetCurrentFolderRepoState returns the RepoState for the currently selected
// folder, or nil if the folder is not a git repository.
func (m *Model) GetCurrentFolderRepoState() *report.RepoState {
	entry := m.GetCurrentFolderEntry()
	if entry == nil || !entry.IsRepo {
		return nil
	}
	rs, ok := m.repoStatesByPath[entry.Path]
	if !ok {
		return nil
	}
	return &rs
}

func (m *Model) DisplayMode() tableDisplayMode {
	return m.displayMode
}
