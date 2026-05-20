package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func reproTheme(t *testing.T) theme.Theme {
	colors, err := theme.CreateColors("")
	if err != nil {
		t.Fatal(err)
	}
	return theme.Theme{Colors: colors, Styles: theme.CreateStyles(colors)}
}

func key(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func reproModel(t *testing.T, vm ViewMode, dirtyCount, cleanCount int) Model {
	th := reproTheme(t)
	full := report.ScanReport{}
	for i := 0; i < dirtyCount; i++ {
		p := t.TempDir()
		full.RepoStates = append(full.RepoStates, report.RepoState{
			ID: "d" + string(rune('0'+i)), Path: p, Repo: "dirty" + string(rune('0'+i)),
			UncommitedFiles: []string{"M x"},
		})
		full.AllFolders = append(full.AllFolders, report.FolderEntry{Path: p, Name: "dirty" + string(rune('0'+i)), IsRepo: true})
	}
	for i := 0; i < cleanCount; i++ {
		p := t.TempDir()
		full.RepoStates = append(full.RepoStates, report.RepoState{
			ID: "c" + string(rune('0'+i)), Path: p, Repo: "clean" + string(rune('0'+i)),
		})
		full.AllFolders = append(full.AllFolders, report.FolderEntry{Path: p, Name: "clean" + string(rune('0'+i)), IsRepo: true})
	}

	m := Model{
		theme:              th,
		fullReport:         full,
		viewMode:           vm,
		reposTable:         repostable.New(th, filterDirtyRepos(full), 80, 10),
		deleteConfirmInput: createDeleteConfirmInputModel(),
		repoDetails:        repodetails.New(nil, th),
		focusStack:         []FocusState{FocusReposTable},
	}
	m.applyViewMode()
	return m
}

func reproDelete(t *testing.T, m Model) Model {
	u, _ := m.Update(key('d'))
	m = u.(Model)
	if m.currentFocus() != FocusDeleteRepoPopup {
		t.Fatalf("d did not open popup; focus=%v path=%q", m.currentFocus(), m.reposTable.GetCurrentPath())
	}
	m.deleteConfirmInput.SetValue("YES")
	u, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = u.(Model)
	msg := cmd()
	u, _ = m.Update(msg)
	return u.(Model)
}

func TestReproScenarios(t *testing.T) {
	for _, vm := range []ViewMode{ViewModeDirty, ViewModeAllRepos, ViewModeAllDirs} {
		m := reproModel(t, vm, 2, 1)
		t.Logf("[vm=%v] before: count=%d path=%q", vm, m.reposTable.ReposCount(), m.reposTable.GetCurrentPath())
		m = reproDelete(t, m)
		t.Logf("[vm=%v] after delete: count=%d cursor=%d path=%q getrepostate=%v",
			vm, m.reposTable.ReposCount(), m.reposTable.Cursor(),
			m.reposTable.GetCurrentPath(), m.reposTable.GetCurrentRepoState() != nil)
		u, _ := m.Update(key('d'))
		m2 := u.(Model)
		t.Logf("[vm=%v] second d focus=%v", vm, m2.currentFocus())
	}
}
