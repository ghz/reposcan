package gitx

import (
	"errors"
	"os"
	"path/filepath"
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

func TestParseTrackCount(t *testing.T) {
	tests := []struct {
		name       string
		track      string
		wantAhead  int
		wantBehind int
	}{
		{name: "empty", track: "", wantAhead: 0, wantBehind: 0},
		{name: "ahead only", track: "[ahead 2]", wantAhead: 2, wantBehind: 0},
		{name: "behind only", track: "[behind 3]", wantAhead: 0, wantBehind: 3},
		{name: "ahead and behind", track: "[ahead 1, behind 4]", wantAhead: 1, wantBehind: 4},
		{name: "gone", track: "[gone]", wantAhead: 0, wantBehind: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTrackCount(tt.track, reTrackAhead); got != tt.wantAhead {
				t.Fatalf("parseTrackCount(ahead) = %d, want %d", got, tt.wantAhead)
			}
			if got := parseTrackCount(tt.track, reTrackBehind); got != tt.wantBehind {
				t.Fatalf("parseTrackCount(behind) = %d, want %d", got, tt.wantBehind)
			}
		})
	}
}

func TestGetBranchStatuses(t *testing.T) {
	dir := t.TempDir()
	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo() error = %v", err)
	}
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := AddAll(dir); err != nil {
		t.Fatalf("AddAll() error = %v", err)
	}
	if err := GitCommit(dir, "initial"); err != nil {
		t.Fatalf("GitCommit() error = %v", err)
	}
	mustGit(t, dir, "branch", "feature-x")

	branches, err := GetBranchStatuses(dir)
	if err != nil {
		t.Fatalf("GetBranchStatuses() error = %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("GetBranchStatuses() returned %d branches, want 2", len(branches))
	}

	byName := map[string]BranchStatus{}
	for _, b := range branches {
		byName[b.Name] = b
	}

	current, ok := byName[currentBranchName(t, dir)]
	if !ok || !current.IsCurrent {
		t.Fatalf("expected current branch to be marked IsCurrent, got %+v", byName)
	}
	for _, b := range branches {
		if b.Upstream != "" {
			t.Fatalf("branch %q has unexpected upstream %q", b.Name, b.Upstream)
		}
		if b.Ahead != 0 || b.Behind != 0 || b.Gone {
			t.Fatalf("branch %q has unexpected divergence %+v", b.Name, b)
		}
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := RunGitCommand(dir, args...); err != nil {
		t.Fatalf("git %s error = %v", strings.Join(args, " "), err)
	}
}

func currentBranchName(t *testing.T, dir string) string {
	t.Helper()
	branch, err := GetRepoBranch(dir)
	if err != nil {
		t.Fatalf("GetRepoBranch() error = %v", err)
	}
	return branch
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
