package repostable

import (
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestCreateRowsFavoriteRepoNameHasNoEmbeddedANSI(t *testing.T) {
	colors, err := theme.CreateColors("catppuccin-mocha")
	if err != nil {
		t.Fatalf("CreateColors() error = %v", err)
	}
	tm := theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}

	rows := createRows([]report.RepoState{
		{
			Repo:   "reposcan",
			Path:   "/tmp/reposcan",
			Branch: "main",
		},
	}, map[string]bool{"/tmp/reposcan": true}, tm)

	if got, want := rows[0][0], "★ reposcan"; got != want {
		t.Fatalf("favorite repo name = %q, want %q", got, want)
	}
	if strings.Contains(rows[0][0], "\x1b[") {
		t.Fatalf("favorite repo name contains ANSI styling: %q", rows[0][0])
	}
}
