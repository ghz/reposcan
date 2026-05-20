// Package repostable is a Model that renders git repo states in a table. Providing functionality like filterning
package repostable

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
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
		favorites:     map[string]bool{},
		expanded:      map[string]bool{},
		branchCache:   map[string][]gitx.BranchStatus{},
	}

	cols := createColumns(width)
	rows, refs := model.buildRepoRows()
	model.rowRefs = refs

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(height),
	)
	t.Focus()

	km := table.DefaultKeyMap()
	setKeymaps(km)

	if len(rows) == 0 {
		t.SetRows([]table.Row{{"", "", "", ""}})
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

// SetFavorites updates the set of pinned repo paths and refreshes the table.
func (m *Model) SetFavorites(paths []string) {
	m.favorites = make(map[string]bool, len(paths))
	for _, p := range paths {
		m.favorites[p] = true
	}
	m.Filter(m.filterQuery)
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

	m.Filter(m.filterQuery)
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
	if m.displayMode == tableDisplayFolders {
		m.filterFolders(query)
		return
	}

	m.filterRepos(query)
}

func (m *Model) filterRepos(query string) {
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

	// Favorites always appear first
	if len(m.favorites) > 0 {
		favs := make([]report.RepoState, 0)
		rest := make([]report.RepoState, 0)
		for _, rs := range m.filteredRepos {
			if m.favorites[rs.Path] {
				favs = append(favs, rs)
			} else {
				rest = append(rest, rs)
			}
		}
		m.filteredRepos = append(favs, rest...)
	}

	cursorPosition := m.tbl.Cursor()

	rows, refs := m.buildRepoRows()
	m.rowRefs = refs
	if len(rows) == 0 {
		rows = []table.Row{{"", "", "", ""}}
	}
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.rowRefs) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}
}

func (m *Model) filterFolders(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	if len(q) == 0 {
		m.filteredFolders = m.folders
	} else {
		m.filteredFolders = []report.FolderEntry{}
		for _, f := range m.folders {
			if strings.Contains(strings.ToLower(f.Name), q) ||
				strings.Contains(strings.ToLower(f.Path), q) {
				m.filteredFolders = append(m.filteredFolders, f)
			}
		}
	}

	cursorPosition := m.tbl.Cursor()

	rows := createFolderRows(m.filteredFolders, m.repoStatesByPath, m.theme)
	if len(rows) == 0 {
		rows = []table.Row{{"", "", "", ""}}
	}
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredFolders) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}
}

// UpdateRepoState replaces the state of the repo identified by newState.ID in
// both the filtered and full repo lists, then rebuilds the table rows.
func (m *Model) UpdateRepoState(_ int, newState report.RepoState) {
	if fi := getRepoIndex(m.filteredRepos, newState.ID); fi != -1 {
		m.filteredRepos[fi] = newState
	}
	if oi := getRepoIndex(m.report.RepoStates, newState.ID); oi != -1 {
		m.report.RepoStates[oi] = newState
	}

	// A refreshed repo may have new branches or a different checked-out
	// branch. Re-fetch immediately when expanded so the visible rows stay
	// correct; otherwise just drop the stale cache.
	if m.expanded[newState.ID] {
		m.branchCache[newState.ID] = fetchBranches(newState.Path)
	} else {
		delete(m.branchCache, newState.ID)
	}

	rows, refs := m.buildRepoRows()
	m.rowRefs = refs
	if len(rows) == 0 {
		rows = []table.Row{{"", "", "", ""}}
	}
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

func (m *Model) SetCursorByRepoID(id string) bool {
	if id == "" || m.displayMode != tableDisplayRepos {
		return false
	}
	for i, ref := range m.rowRefs {
		if ref.branchIx == -1 && m.filteredRepos[ref.repoIdx].ID == id {
			m.tbl.SetCursor(i)
			return true
		}
	}
	return false
}

// ReposCount returns the number of selectable rows: folders in folder mode, or
// visible rows (repo headers plus any expanded branch rows) in repos mode.
func (rt *Model) ReposCount() int {
	if rt.displayMode == tableDisplayFolders {
		return len(rt.filteredFolders)
	}
	return len(rt.rowRefs)
}

// currentRef returns the rowRef under the cursor in repos mode.
func (m *Model) currentRef() (rowRef, bool) {
	if m.displayMode != tableDisplayRepos {
		return rowRef{}, false
	}
	i := m.tbl.Cursor()
	if i < 0 || i >= len(m.rowRefs) {
		return rowRef{}, false
	}
	return m.rowRefs[i], true
}

// CurrentRowIsBranch reports whether the cursor is on a branch (child) row.
func (m *Model) CurrentRowIsBranch() bool {
	ref, ok := m.currentRef()
	return ok && ref.branchIx >= 0
}

// GetCurrentBranchStatus returns the branch under the cursor when the cursor
// is on a branch (child) row, or nil otherwise.
func (m *Model) GetCurrentBranchStatus() *gitx.BranchStatus {
	ref, ok := m.currentRef()
	if !ok || ref.branchIx < 0 {
		return nil
	}
	rs := m.filteredRepos[ref.repoIdx]
	branches := m.branchCache[rs.ID]
	if ref.branchIx >= len(branches) {
		return nil
	}
	return &branches[ref.branchIx]
}

// ExpandCurrent expands the repo under the cursor, fetching its branch list
// the first time. No-op in folder mode or when already expanded.
func (m *Model) ExpandCurrent() {
	m.setExpanded(true)
}

// CollapseCurrent collapses the repo under the cursor (or the parent repo when
// the cursor sits on a branch row).
func (m *Model) CollapseCurrent() {
	m.setExpanded(false)
}

func (m *Model) setExpanded(expand bool) {
	ref, ok := m.currentRef()
	if !ok {
		return
	}
	rs := m.filteredRepos[ref.repoIdx]
	if m.expanded[rs.ID] == expand {
		return
	}

	if expand {
		if _, cached := m.branchCache[rs.ID]; !cached {
			m.branchCache[rs.ID] = fetchBranches(rs.Path)
		}
	}
	m.expanded[rs.ID] = expand

	rows, refs := m.buildRepoRows()
	m.rowRefs = refs
	if len(rows) == 0 {
		rows = []table.Row{{"", "", "", ""}}
	}
	m.tbl.SetRows(rows)
	m.SetCursorByRepoID(rs.ID)
}

func fetchBranches(path string) []gitx.BranchStatus {
	branches, err := gitx.GetBranchStatuses(path)
	if err != nil {
		return []gitx.BranchStatus{}
	}
	return branches
}

func (m *Model) GetCurrentRepoState() *report.RepoState {
	if m.displayMode == tableDisplayFolders {
		return m.GetCurrentFolderRepoState()
	}
	return m.GetRepoStateAt(m.Cursor())
}

// GetRepoStateAt returns the repo for the given visual row index. Branch rows
// resolve to their parent repo.
func (m *Model) GetRepoStateAt(index int) *report.RepoState {
	if index < 0 || index >= len(m.rowRefs) {
		return nil
	}
	ri := m.rowRefs[index].repoIdx
	if ri < 0 || ri >= len(m.filteredRepos) {
		return nil
	}
	return &m.filteredRepos[ri]
}

// GetCurrentPath returns the filesystem path of the currently selected item,
// regardless of whether the table is in repos or folders mode.
func (m *Model) GetCurrentPath() string {
	if m.displayMode == tableDisplayFolders {
		entry := m.GetCurrentFolderEntry()
		if entry == nil {
			return ""
		}
		return entry.Path
	}
	rs := m.GetCurrentRepoState()
	if rs == nil {
		return ""
	}
	return rs.Path
}

// GetCurrentFolderEntry returns the FolderEntry at the current cursor in folders mode.
func (m *Model) GetCurrentFolderEntry() *report.FolderEntry {
	if m.displayMode != tableDisplayFolders {
		return nil
	}
	i := m.tbl.Cursor()
	if i < 0 || i >= len(m.filteredFolders) {
		return nil
	}
	return &m.filteredFolders[i]
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
