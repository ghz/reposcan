package repostable

import (
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func testTheme(t *testing.T) theme.Theme {
	t.Helper()
	colors, err := theme.CreateColors("catppuccin-mocha")
	if err != nil {
		t.Fatalf("CreateColors() error = %v", err)
	}
	return theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}
}

func TestRepoRowFavoriteNameHasNoEmbeddedANSI(t *testing.T) {
	row := repoRow(report.RepoState{
		Repo:   "reposcan",
		Path:   "/tmp/reposcan",
		Branch: "main",
	}, false, true, testTheme(t))

	if got, want := row[0], "▸ ★ reposcan"; got != want {
		t.Fatalf("favorite repo name = %q, want %q", got, want)
	}
	if strings.Contains(row[0], "\x1b[") {
		t.Fatalf("favorite repo name contains ANSI styling: %q", row[0])
	}
}

func TestRepoRowMarkerReflectsExpandedState(t *testing.T) {
	tm := testTheme(t)
	rs := report.RepoState{Repo: "reposcan", Path: "/tmp/reposcan", Branch: "main"}

	if got := repoRow(rs, false, false, tm)[0]; !strings.HasPrefix(got, "▸ ") {
		t.Fatalf("collapsed marker = %q, want ▸ prefix", got)
	}
	if got := repoRow(rs, true, false, tm)[0]; !strings.HasPrefix(got, "▾ ") {
		t.Fatalf("expanded marker = %q, want ▾ prefix", got)
	}
}

func TestBuildRepoRowsExpandsBranches(t *testing.T) {
	m := New(testTheme(t), report.ScanReport{RepoStates: []report.RepoState{
		{ID: "a", Repo: "alpha", Path: "/a", Branch: "main"},
		{ID: "b", Repo: "beta", Path: "/b", Branch: "dev"},
	}}, 100, 10)

	rows, refs := m.buildRepoRows()
	if len(rows) != 2 || len(refs) != 2 {
		t.Fatalf("collapsed: rows=%d refs=%d, want 2/2", len(rows), len(refs))
	}

	m.expanded["a"] = true
	m.branchCache["a"] = []gitx.BranchStatus{
		{Name: "main", IsCurrent: true, Upstream: "origin/main", Ahead: 1},
		{Name: "feature", Upstream: "origin/feature", Behind: 2},
	}

	rows, refs = m.buildRepoRows()
	if len(rows) != 4 {
		t.Fatalf("expanded: rows=%d, want 4 (repo a + 2 branches + repo b)", len(rows))
	}
	want := []rowRef{
		{repoIdx: 0, branchIx: -1},
		{repoIdx: 0, branchIx: 0},
		{repoIdx: 0, branchIx: 1},
		{repoIdx: 1, branchIx: -1},
	}
	for i, w := range want {
		if refs[i] != w {
			t.Fatalf("refs[%d] = %+v, want %+v", i, refs[i], w)
		}
	}
}

func TestBranchRowResolvesToParentRepo(t *testing.T) {
	m := New(testTheme(t), report.ScanReport{RepoStates: []report.RepoState{
		{ID: "a", Repo: "alpha", Path: "/a", Branch: "main"},
	}}, 100, 10)

	m.expanded["a"] = true
	m.branchCache["a"] = []gitx.BranchStatus{
		{Name: "main", IsCurrent: true},
		{Name: "feature"},
	}
	rows, refs := m.buildRepoRows()
	m.rowRefs = refs
	m.tbl.SetRows(rows)

	// Cursor on the second branch row (visual index 2).
	m.tbl.SetCursor(2)
	if !m.CurrentRowIsBranch() {
		t.Fatal("CurrentRowIsBranch() = false, want true on a branch row")
	}
	rs := m.GetCurrentRepoState()
	if rs == nil || rs.ID != "a" {
		t.Fatalf("branch row resolved to %+v, want parent repo a", rs)
	}
}
