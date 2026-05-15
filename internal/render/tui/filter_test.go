package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestReposFilterEnterClosesFilterAndKeepsSelectedRepo(t *testing.T) {
	m := newTestModelWithRepos(t, []report.RepoState{
		{ID: "1", Path: "/repos/core", Repo: "core", Branch: "main"},
		{ID: "2", Path: "/repos/api", Repo: "api-service", Branch: "main"},
		{ID: "3", Path: "/repos/web", Repo: "web", Branch: "main"},
	})

	m.pushFocus(FocusReposFilter)
	m.reposFilter.SetValue("api")
	m.reposTable.Filter("api")

	updated, _ := m.updateReposFilter(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.currentFocus() != FocusReposTable {
		t.Fatalf("focus = %v, want %v", got.currentFocus(), FocusReposTable)
	}
	if got.IsReposFilterVisible() {
		t.Fatal("filter is still visible after enter")
	}

	rs := got.reposTable.GetCurrentRepoState()
	if rs == nil {
		t.Fatal("selected repo = nil, want api-service")
	}
	if rs.ID != "2" {
		t.Fatalf("selected repo ID = %q, want %q", rs.ID, "2")
	}
}

func TestReposFilterNoMatchesCanBeAccepted(t *testing.T) {
	m := newTestModelWithRepos(t, []report.RepoState{
		{ID: "1", Path: "/repos/core", Repo: "core", Branch: "main"},
	})

	m.pushFocus(FocusReposFilter)
	m.reposFilter.SetValue("zzz")
	m.reposTable.Filter("zzz")

	updated, _ := m.updateReposFilter(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.currentFocus() != FocusReposTable {
		t.Fatalf("focus = %v, want %v", got.currentFocus(), FocusReposTable)
	}
	if got.IsReposFilterVisible() {
		t.Fatal("filter is still visible after enter")
	}
	if rs := got.reposTable.GetCurrentRepoState(); rs == nil || rs.ID != "1" {
		t.Fatalf("selected repo after clearing no-match filter = %#v, want repo ID 1", rs)
	}
}

func TestReposFilterFiltersCurrentFolderView(t *testing.T) {
	m := newTestModelWithRepos(t, nil)
	m.reposTable.SetFolders([]report.FolderEntry{
		{Name: "alpha", Path: "/repos/alpha"},
		{Name: "beta-tools", Path: "/repos/beta-tools"},
		{Name: "gamma", Path: "/repos/gamma"},
	}, nil)

	m.pushFocus(FocusReposFilter)
	m.reposFilter.SetValue("beta")
	m.reposTable.Filter("beta")

	if got := m.reposTable.ReposCount(); got != 1 {
		t.Fatalf("filtered folder count = %d, want 1", got)
	}
	entry := m.reposTable.GetCurrentFolderEntry()
	if entry == nil || entry.Name != "beta-tools" {
		t.Fatalf("selected folder after filtering = %#v, want beta-tools", entry)
	}

	updated, _ := m.updateReposFilter(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.IsReposFilterVisible() {
		t.Fatal("filter is still visible after enter")
	}
	if count := got.reposTable.ReposCount(); count != 3 {
		t.Fatalf("folder count after accepting filter = %d, want 3", count)
	}
	if entry := got.reposTable.GetCurrentFolderEntry(); entry == nil || entry.Name != "alpha" {
		t.Fatalf("selected folder after clearing filter = %#v, want alpha", entry)
	}
}

func TestReposTableFooterKeybindingsRouteThroughUpdate(t *testing.T) {
	tests := []struct {
		name      string
		key       tea.KeyMsg
		wantFocus FocusState
		wantCmd   bool
	}{
		{
			name:      "git menu",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}},
			wantFocus: FocusGitMenuPopup,
		},
		{
			name:      "git menu uppercase",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}},
			wantFocus: FocusGitMenuPopup,
		},
		{
			name:      "open remote",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}},
			wantFocus: FocusReposTable,
			wantCmd:   true,
		},
		{
			name:      "delete popup",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			wantFocus: FocusDeleteRepoPopup,
		},
		{
			name:      "search",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			wantFocus: FocusReposFilter,
		},
		{
			name:      "help",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			wantFocus: FocusHelpPopup,
		},
		{
			name:      "refresh",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			wantFocus: FocusReposTable,
			wantCmd:   true,
		},
		{
			name:      "copy path",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
			wantFocus: FocusReposTable,
			wantCmd:   true,
		},
		{
			name:      "quit",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantFocus: FocusReposTable,
			wantCmd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelWithRepos(t, []report.RepoState{{
				ID:     "1",
				Path:   t.TempDir(),
				Repo:   "repo",
				Branch: "main",
			}})

			updated, cmd := m.Update(tt.key)
			got := updated.(Model)

			if got.currentFocus() != tt.wantFocus {
				t.Fatalf("focus = %v, want %v", got.currentFocus(), tt.wantFocus)
			}
			if (cmd != nil) != tt.wantCmd {
				t.Fatalf("cmd nil = %v, want command presence %v", cmd == nil, tt.wantCmd)
			}
		})
	}
}

func TestFocusStackDefaultsToReposTableWhenEmpty(t *testing.T) {
	m := newTestModelWithRepos(t, []report.RepoState{{
		ID:     "1",
		Path:   t.TempDir(),
		Repo:   "repo",
		Branch: "main",
	}})
	m.focusStack = nil

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	got := updated.(Model)
	if got.currentFocus() != FocusReposFilter {
		t.Fatalf("focus after / = %v, want %v", got.currentFocus(), FocusReposFilter)
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got = updated.(Model)
	if got.currentFocus() != FocusReposTable {
		t.Fatalf("focus after esc = %v, want %v", got.currentFocus(), FocusReposTable)
	}
}

func newTestModelWithRepos(t *testing.T, repos []report.RepoState) Model {
	t.Helper()

	colors, err := theme.CreateColors("")
	if err != nil {
		t.Fatalf("CreateColors() error = %v", err)
	}
	th := theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}
	scanReport := report.ScanReport{RepoStates: repos}

	return Model{
		reposTable:         repostable.New(th, scanReport, 100, 10),
		alerts:             alerts.New(th),
		reposFilter:        createRrepoFilter(),
		deleteConfirmInput: createDeleteConfirmInputModel(),
		theme:              th,
		focusStack:         []FocusState{FocusReposTable},
	}
}
