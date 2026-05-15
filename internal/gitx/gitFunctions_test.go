package gitx

import (
	"errors"
	"strings"
	"testing"
)

func TestCommitInitialIfNeededSkipsEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}

	committed, err := CommitInitialIfNeeded(dir)
	if err != nil {
		t.Fatalf("CommitInitialIfNeeded() error = %v", err)
	}
	if committed {
		t.Fatal("CommitInitialIfNeeded() committed = true, want false")
	}
}

func TestIsNothingToCommitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nothing to commit",
			err:  errors.New("exit status 1: On branch main\nnothing to commit, working tree clean"),
			want: true,
		},
		{
			name: "no changes added",
			err:  errors.New("exit status 1: no changes added to commit"),
			want: true,
		},
		{
			name: "other git error",
			err:  errors.New("exit status 128: repository not found"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNothingToCommitError(tt.err); got != tt.want {
				t.Fatalf("isNothingToCommitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHubCreateRepoRequiresName(t *testing.T) {
	err := GitHubCreateRepo("   ", t.TempDir(), true, false)
	if err == nil {
		t.Fatal("GitHubCreateRepo() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Fatalf("GitHubCreateRepo() error = %q, want empty-name message", err.Error())
	}
}
