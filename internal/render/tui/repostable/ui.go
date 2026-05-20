package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/pkg/report"
)

const (
	RepoW        = 25
	BranchW      = 25
	RemoteStateW = 35
	LastCommitW  = 15
)

func createColumns(maxWidth int, sortKey SortKey, sortAsc bool) []table.Column {
	repoW := maxWidth * RepoW / 100
	branchW := maxWidth * BranchW / 100
	remoteStateW := maxWidth * RemoteStateW / 100
	lastCommitW := maxWidth * LastCommitW / 100

	titles := []string{"Repo", "Branch", "Last Commit", "State"}
	arrow := " ▲"
	if !sortAsc {
		arrow = " ▼"
	}
	if int(sortKey) >= 0 && int(sortKey) < len(titles) {
		titles[sortKey] += arrow
	}

	return []table.Column{
		{Title: titles[SortByRepo], Width: repoW},
		{Title: titles[SortByBranch], Width: branchW},
		{Title: titles[SortByLastCommit], Width: lastCommitW},
		{Title: titles[SortByState], Width: remoteStateW},
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

	lastCommitStr := utils.RelativeTime(rs.LastCommitTime)

	return table.Row{
		name,
		rs.Branch,
		preserveBackground(t.Styles.Muted.Render(lastCommitStr)),
		preserveBackground(getStateColumnStr(rs, t)),
	}
}

// preserveBackground replaces lipgloss's full ANSI reset (\x1b[0m) with a
// foreground-only reset (\x1b[39m) so that an outer background (e.g. the
// selected-row highlight applied by bubbles/table) is not cleared by the
// per-token styling inside a cell.
func preserveBackground(s string) string {
	return strings.ReplaceAll(s, "\x1b[0m", "\x1b[39m")
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
		preserveBackground(upstream),
		preserveBackground(t.Styles.Muted.Render("—")),
		preserveBackground(getBranchStateStr(b, t)),
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
		lastCommit := t.Styles.Muted.Render("—")

		if f.IsRepo {
			typeLabel = t.Styles.Base.Foreground(t.Colors.Info).Render("repo")
			if rs, ok := reposByPath[f.Path]; ok {
				branch = rs.Branch
				state = getStateColumnStr(rs, t)
				lastCommit = t.Styles.Muted.Render(utils.RelativeTime(rs.LastCommitTime))
			}
		}

		rows = append(rows, table.Row{
			preserveBackground(f.Name + "  " + typeLabel),
			preserveBackground(branch),
			preserveBackground(lastCommit),
			preserveBackground(state),
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
