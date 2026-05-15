package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

const (
	RepoW        = 30
	BranchW      = 30
	RemoteStateW = 40
)

func createColumns(maxWidth int) []table.Column {
	repoW := maxWidth * RepoW / 100
	branchW := maxWidth * BranchW / 100
	remoteStateW := maxWidth * RemoteStateW / 100

	return []table.Column{
		{Title: "Repo", Width: repoW},
		{Title: "Branch", Width: branchW},
		{Title: "State", Width: remoteStateW},
	}
}

// buildRepoRows turns filteredRepos into table rows. Expanded repos are
// followed by one indented row per local branch. The returned rowRefs slice is
// parallel to the rows and maps each visual row back to its repo (and branch).
func (m *Model) buildRepoRows() ([]table.Row, []rowRef) {
	rows := make([]table.Row, 0, len(m.filteredRepos))
	refs := make([]rowRef, 0, len(m.filteredRepos))

	for i, rs := range m.filteredRepos {
		expanded := m.expanded[rs.ID]
		rows = append(rows, repoRow(rs, expanded, m.favorites[rs.Path], m.theme))
		refs = append(refs, rowRef{repoIdx: i, branchIx: -1})

		if expanded {
			for bi, b := range m.branchCache[rs.ID] {
				rows = append(rows, branchRow(b, m.theme))
				refs = append(refs, rowRef{repoIdx: i, branchIx: bi})
			}
		}
	}
	return rows, refs
}

// repoRow renders a single repo header row. The leading marker shows whether
// the repo is expanded (▾) or collapsed (▸).
func repoRow(rs report.RepoState, expanded, favorite bool, t theme.Theme) table.Row {
	marker := "▸ "
	if expanded {
		marker = "▾ "
	}

	name := marker
	if favorite {
		name += "★ "
	}
	name += rs.Repo

	return table.Row{
		name,
		rs.Branch,
		getStateColumnStr(rs, t),
	}
}

// branchRow renders one indented branch row beneath its repo. The current
// branch is marked with a "*".
func branchRow(b gitx.BranchStatus, t theme.Theme) table.Row {
	indent := "    "
	if b.IsCurrent {
		indent = "  * "
	}

	upstream := t.Styles.Muted.Render("—")
	if b.Upstream != "" {
		upstream = t.Styles.Muted.Render("→ " + b.Upstream)
	}

	return table.Row{
		indent + b.Name,
		upstream,
		getBranchStateStr(b, t),
	}
}

// getBranchStateStr renders the ahead/behind status for a single branch.
func getBranchStateStr(b gitx.BranchStatus, t theme.Theme) string {
	switch {
	case b.Upstream == "":
		return t.Styles.Muted.Render("no upstream")
	case b.Gone:
		return lipgloss.NewStyle().Foreground(t.Colors.Error).Render("upstream gone")
	default:
		aheadColor := t.Colors.Muted
		if b.Ahead > 0 {
			aheadColor = t.Colors.Warning
		}
		behindColor := t.Colors.Muted
		if b.Behind > 0 {
			behindColor = t.Colors.Warning
		}
		ahead := lipgloss.NewStyle().Foreground(aheadColor).Render(fmt.Sprintf("↑%-d", b.Ahead))
		behind := lipgloss.NewStyle().Foreground(behindColor).Render(fmt.Sprintf("↓%-d", b.Behind))
		return ahead + " " + behind
	}
}

func getStateColumnStr(rs report.RepoState, t theme.Theme) string {
	uc := len(rs.UncommitedFiles)

	ucColor := t.Colors.Muted
	if uc > 0 {
		ucColor = t.Colors.Error
	}
	ucStr := lipgloss.NewStyle().Foreground(ucColor).Render(fmt.Sprintf("⏳%-d", uc)) + " "

	parts := []string{}
	for _, remoteStatus := range rs.RemoteStatus {
		var statusParts []string

		if remoteStatus.Ahead < 0 {
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(t.Colors.Error).Render("x"))
		} else {
			aheadColor := t.Colors.Muted
			if remoteStatus.Ahead > 0 {
				aheadColor = t.Colors.Warning
			}
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(aheadColor).Render(fmt.Sprintf("↑%-d", remoteStatus.Ahead)))
		}

		if remoteStatus.Behind < 0 {
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(t.Colors.Error).Render("x"))
		} else {
			behindColor := t.Colors.Muted
			if remoteStatus.Behind > 0 {
				behindColor = t.Colors.Warning
			}
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(behindColor).Render(fmt.Sprintf("↓%-d", remoteStatus.Behind)))
		}

		if remoteStatus.Remote != "" && !(len(rs.RemoteStatus) == 1 && remoteStatus.Remote == "origin") {
			remoteName := t.Styles.Base.Render(fmt.Sprintf("(%s)", remoteStatus.Remote))
			statusParts = append(statusParts, remoteName)
		}

		parts = append(parts, strings.Join(statusParts, " "))
	}

	s := ucStr
	s += strings.Join(parts, " | ")

	return s
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", tea.KeyCtrlU.String())
	km.PageDown.SetKeys("pgdn", tea.KeyCtrlD.String())
	km.GotoTop.SetKeys("home")
	km.GotoBottom.SetKeys("end")
}

func createFolderRows(folders []report.FolderEntry, reposByPath map[string]report.RepoState, t theme.Theme) []table.Row {
	rows := make([]table.Row, 0, len(folders))
	for _, f := range folders {
		typeLabel := t.Styles.Muted.Render("dir")
		branch := t.Styles.Muted.Render("—")
		state := t.Styles.Muted.Render("—")

		if f.IsRepo {
			typeLabel = t.Styles.Base.Foreground(t.Colors.Info).Render("repo")
			if rs, ok := reposByPath[f.Path]; ok {
				branch = rs.Branch
				state = getStateColumnStr(rs, t)
			}
		}

		rows = append(rows, table.Row{
			f.Name + "  " + typeLabel,
			branch,
			state,
		})
	}
	return rows
}

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
