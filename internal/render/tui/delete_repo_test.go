package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestDeleteRepoCmdRemovesDirectory(t *testing.T) {
	dir := t.TempDir()
	repoPath := dir + string(os.PathSeparator) + "repo"
	if err := os.Mkdir(repoPath, 0o755); err != nil {
		t.Fatalf("Mkdir(repo) error = %v", err)
	}

	msg := deleteRepoCmd("repo", repoPath)()
	result, ok := msg.(deleteRepoResultMsg)
	if !ok {
		t.Fatalf("message = %T, want deleteRepoResultMsg", msg)
	}
	if result.err != nil {
		t.Fatalf("deleteRepoCmd() error = %v", result.err)
	}
	if _, statErr := os.Lstat(repoPath); !os.IsNotExist(statErr) {
		t.Fatalf("repo directory still exists or unexpected stat error: %v", statErr)
	}
}

func TestDeleteRepoCmdTimesOutWhenRemoveBlocks(t *testing.T) {
	oldRemoveAll := removeAll
	oldTimeout := deleteRepoTimeout
	removeAll = func(string) error {
		select {}
	}
	deleteRepoTimeout = 10 * time.Millisecond
	t.Cleanup(func() {
		removeAll = oldRemoveAll
		deleteRepoTimeout = oldTimeout
	})

	msg := deleteRepoCmd("repo", t.TempDir())()
	result, ok := msg.(deleteRepoResultMsg)
	if !ok {
		t.Fatalf("message = %T, want deleteRepoResultMsg", msg)
	}
	if result.err == nil {
		t.Fatal("deleteRepoCmd() error = nil, want timeout error")
	}
	if !strings.Contains(result.err.Error(), "timed out") {
		t.Fatalf("deleteRepoCmd() error = %q, want timeout", result.err)
	}
}

func TestDeleteRepoPopupRequiresYES(t *testing.T) {
	m := testDeleteRepoModel(t)
	m.deleteConfirmInput.SetValue("yes")

	updated, cmd := m.updateDeleteRepoPopup(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("updateDeleteRepoPopup() cmd = nil, want alert command")
	}

	got := updated.(Model)
	if got.currentFocus() != FocusDeleteRepoPopup {
		t.Fatalf("focus = %v, want FocusDeleteRepoPopup", got.currentFocus())
	}
}

func TestDeleteRepoFailureReturnsToTableWithoutLoading(t *testing.T) {
	m := testDeleteRepoModel(t)
	m.loading = true

	updated, cmd := m.Update(deleteRepoResultMsg{
		repoName: "repo",
		path:     m.reposTable.GetCurrentPath(),
		err:      errors.New("locked"),
	})
	got := updated.(Model)

	if got.currentFocus() != FocusReposTable {
		t.Fatalf("focus = %v, want FocusReposTable", got.currentFocus())
	}
	if got.loading {
		t.Fatal("loading = true, want false after failed delete")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want error alert command")
	}
}

func TestDeleteRepoSuccessRemovesPathWithoutLoadingOrRefresh(t *testing.T) {
	m := testDeleteRepoModel(t)
	repoPath := m.reposTable.GetCurrentPath()
	otherPath := t.TempDir()
	m.fullReport = report.ScanReport{
		RepoStates: []report.RepoState{
			{ID: "repo-id", Path: repoPath, Repo: "repo"},
			{ID: "other-id", Path: otherPath, Repo: "other"},
		},
		AllFolders: []report.FolderEntry{
			{Path: repoPath, Name: "repo", IsRepo: true},
			{Path: otherPath, Name: "other", IsRepo: true},
		},
	}

	updated, cmd := m.Update(deleteRepoResultMsg{
		repoName: "repo",
		path:     repoPath,
	})
	got := updated.(Model)

	if got.currentFocus() != FocusReposTable {
		t.Fatalf("focus = %v, want FocusReposTable", got.currentFocus())
	}
	if got.loading {
		t.Fatal("loading = true, want false after successful delete")
	}
	if len(got.fullReport.RepoStates) != 1 || got.fullReport.RepoStates[0].Path != otherPath {
		t.Fatalf("repo states after delete = %#v, want only %q", got.fullReport.RepoStates, otherPath)
	}
	if len(got.fullReport.AllFolders) != 1 || got.fullReport.AllFolders[0].Path != otherPath {
		t.Fatalf("folders after delete = %#v, want only %q", got.fullReport.AllFolders, otherPath)
	}

	msg := cmd()
	if _, ok := msg.(alerts.AddAlertMsg); !ok {
		t.Fatalf("delete success command = %T, want alerts.AddAlertMsg", msg)
	}
}

func TestDeleteFolderFromNonGitFolderSelection(t *testing.T) {
	m := testDeleteRepoModel(t)
	folderPath := t.TempDir()
	m.reposTable.SetFolders([]report.FolderEntry{{
		Path:   folderPath,
		Name:   "plain-folder",
		IsRepo: false,
	}}, nil)

	updated, cmd := m.updateReposTable(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd != nil {
		t.Fatalf("updateReposTable(d) cmd = %v, want nil before confirmation", cmd)
	}

	got := updated.(Model)
	if got.currentFocus() != FocusDeleteRepoPopup {
		t.Fatalf("focus = %v, want FocusDeleteRepoPopup", got.currentFocus())
	}

	name, path, rs := got.deleteTarget()
	if name != "plain-folder" {
		t.Fatalf("delete target name = %q, want plain-folder", name)
	}
	if path != folderPath {
		t.Fatalf("delete target path = %q, want %q", path, folderPath)
	}
	if rs != nil {
		t.Fatalf("delete target repo state = %#v, want nil", rs)
	}
}

func TestDeleteRepoStatusTextIncludesPushPullAndUncommitted(t *testing.T) {
	rs := report.RepoState{
		UncommitedFiles: []string{"M file.go", "?? new.go"},
		RemoteStatus: []report.RemoteStatus{
			{Remote: "origin", Ahead: 2, Behind: 1},
			{Remote: "backup", Ahead: 3, Behind: 0},
		},
	}

	got := deleteRepoStatusText(&rs)
	for _, want := range []string{
		"2 uncommitted file(s)",
		"5 commit(s) to push",
		"1 commit(s) to pull",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("deleteRepoStatusText() = %q, want substring %q", got, want)
		}
	}
}

func testDeleteRepoModel(t *testing.T) Model {
	t.Helper()

	colors, err := theme.CreateColors("")
	if err != nil {
		t.Fatalf("CreateColors() error = %v", err)
	}
	th := theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}
	r := report.ScanReport{
		RepoStates: []report.RepoState{{
			ID:   "repo-id",
			Path: t.TempDir(),
			Repo: "repo",
		}},
	}

	m := Model{
		theme:              th,
		reposTable:         repostable.New(th, r, 80, 10),
		deleteConfirmInput: createDeleteConfirmInputModel(),
		focusStack:         []FocusState{FocusReposTable, FocusDeleteRepoPopup},
	}
	m.deleteConfirmInput.Focus()
	return m
}
