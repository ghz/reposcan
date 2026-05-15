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
		reposTable:  repostable.New(th, scanReport, 100, 10),
		alerts:      alerts.New(th),
		reposFilter: createRrepoFilter(),
		theme:       th,
		focusStack:  []FocusState{FocusReposTable},
	}
}
