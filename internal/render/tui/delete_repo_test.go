package tui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
