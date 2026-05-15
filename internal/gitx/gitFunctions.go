package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// RemoteStatus holds the ahead/behind status for a specific remote branch
type remoteStatus struct {
	Ahead  int
	Behind int
}

// GitPush pushed git repo at given path using `git push` command and returns stdout of the command + error if any
func GitPush(path string) (string, error) {
	str, err := RunGitCommand(path, "push", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

func GitPull(path string) (string, error) {
	str, err := RunGitCommand(path, "pull")
	if err != nil {
		return "", err
	}
	return str, nil
}

func GitFetch(path string) (string, error) {
	str, err := RunGitCommand(path, "fetch", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

func GetGitRemotes(path string) (remotes []string, err error) {
	str, err := RunGitCommand(path, "remote")
	if err != nil {
		return []string{}, err
	}

	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return []string{}, nil
	}

	remotes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	return remotes, nil
}

// GetRepoBranch returns the current branch name for the Git repository at path.
func GetRepoBranch(path string) (branchName string, err error) {
	str, err := RunGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "-", err
	}
	return strings.TrimSpace(str), nil
}

// GetUncommitedFiles returns the list of uncommitted files (status porcelain)
// for the Git repository at path.
func GetUncommitedFiles(path string) (changes []string, err error) {
	str, err := RunGitCommand(path, "status", "--porcelain=v1", "-uall")
	if err != nil {
		return []string{}, err
	}

	changes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	changes = removeEmptyStrings(changes)

	return changes, nil
}

// GetUpstreamStatus returns the ahead/behind counts relative to the upstream
// tracking branch for the repository at path.
func GetUpstreamStatus(path string) (ahead int, behind int, err error) {
	lrc, err := RunGitCommand(path, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if err != nil {
		return -1, -1, err
	}
	parts := strings.Fields(strings.TrimSpace(lrc))
	if len(parts) == 2 {
		behind = atoiSafe(parts[0])
		ahead = atoiSafe(parts[1])
	}

	return ahead, behind, nil
}

// GetUpstreamStatusForAllRemotes returns the ahead/behind counts for the current branch
// against the same branch on each remote. Returns a slice of RemoteStatus.
func GetUpstreamStatusForAllRemotes(
	path string,
	remote string,
	currentBranch string,
) (remoteStatus, error) {

	// Construct remote branch ref: remote/branch
	remoteBranchRef := remote + "/" + currentBranch

	// Check if remote branch exists
	_, err := RunGitCommand(path, "rev-parse", "--verify", remoteBranchRef)
	if err != nil {
		// Remote branch doesn't exist, skip this remote
		return remoteStatus{}, err
	}

	// Get ahead/behind count for this remote branch
	lrc, err := RunGitCommand(path, "rev-list", "--left-right", "--count", remoteBranchRef+"...HEAD")
	if err != nil {
		return remoteStatus{}, err
	}

	parts := strings.Fields(strings.TrimSpace(lrc))
	var ahead, behind int
	if len(parts) == 2 {
		behind = atoiSafe(parts[0])
		ahead = atoiSafe(parts[1])
	}

	return remoteStatus{
		Ahead:  ahead,
		Behind: behind,
	}, nil
}

// GetRepoName tries to extract the repository name from its remote URL,
// falling back to the first remote name or the local folder name if needed.
func GetRepoName(repoPath string) (string, error) {
	// 1. Try "origin" first
	remote, err := RunGitCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		// 2. If "origin" not found, list remotes
		remotes, rErr := RunGitCommand(repoPath, "remote")
		if rErr == nil {
			names := strings.Fields(remotes)
			if len(names) > 0 {
				remote, err = RunGitCommand(repoPath, "remote", "get-url", names[0])
				if err != nil {
					remote = ""
				}
			}
		}
	}

	remote = strings.TrimSpace(remote)
	if remote != "" {
		if name, ok := parseRepoName(remote); ok {
			return name, nil
		}
	}

	// 3. Fallback to repo folder name
	top, err := RunGitCommand(repoPath, "rev-parse", "--show-toplevel")
	if err == nil {
		return filepath.Base(strings.TrimSpace(top)), nil
	}

	return "", errors.New("could not determine repo name")
}

// parseRepoName extracts the repo name from a remote URL or path.
func parseRepoName(remote string) (string, bool) {
	// handle scp-like: git@host:org/repo.git
	if strings.Contains(remote, ":") && strings.Contains(remote, "@") && !strings.Contains(remote, "://") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) == 2 {
			remote = "ssh://" + parts[0] + "/" + parts[1]
		}
	}

	if u, err := url.Parse(remote); err == nil && u.Path != "" {
		base := path.Base(u.Path)
		base = strings.TrimSuffix(base, ".git")
		return base, true
	}

	// fallback regex
	re := regexp.MustCompile(`([^/\\]+?)(?:\.git)?[/\\]?$`)
	if match := re.FindStringSubmatch(remote); len(match) > 1 {
		return match[1], true
	}

	return "", false
}

// InitRepo initialises a new Git repository at path.
func InitRepo(path string) error {
	_, err := RunGitCommand(path, "init")
	return err
}

// AddAll stages all files in the repository at path.
func AddAll(path string) error {
	_, err := RunGitCommand(path, "add", ".")
	return err
}

// GitCommit creates a commit with the given message in the repository at path.
func GitCommit(path, message string) error {
	_, err := RunGitCommand(path, "commit", "-m", message)
	return err
}

// CommitInitial creates the first commit in the repository at path.
func CommitInitial(path string) error {
	return GitCommit(path, "initial commit")
}

// CommitInitialIfNeeded creates an initial commit when there are staged changes.
// It returns false when the repository has nothing to commit.
func CommitInitialIfNeeded(path string) (bool, error) {
	stagedFiles, err := RunGitCommand(path, "diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(stagedFiles) == "" {
		return false, nil
	}

	if err := CommitInitial(path); err != nil {
		if isNothingToCommitError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isNothingToCommitError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nothing to commit") ||
		strings.Contains(msg, "no changes added to commit")
}

// QuickSaveResult describes what QuickSave did.
type QuickSaveResult struct {
	Committed bool
	Pushed    bool
}

// QuickSave runs git add . → git commit -m "wip" → git push.
// If there is nothing to commit the commit step is skipped and push still runs.
func QuickSave(path string) (QuickSaveResult, error) {
	var r QuickSaveResult

	if err := AddAll(path); err != nil {
		return r, err
	}

	if err := GitCommit(path, "wip"); err != nil {
		if !strings.Contains(err.Error(), "nothing to commit") {
			return r, err
		}
	} else {
		r.Committed = true
	}

	if _, err := GitPush(path); err != nil {
		return r, err
	}
	r.Pushed = true

	return r, nil
}

// GitHubCreateRepo uses the gh CLI to create a GitHub repo named repoName,
// linked to the local directory at path, and optionally pushes the initial commit.
// Requires the gh CLI to be installed and authenticated.
func GitHubCreateRepo(repoName, path string, private bool, push bool) error {
	if strings.TrimSpace(repoName) == "" {
		return errors.New("GitHub repository name cannot be empty")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New("GitHub CLI (gh) is not installed or not in PATH; install it and run `gh auth login`")
	}

	visibility := "--public"
	if private {
		visibility = "--private"
	}
	args := []string{"repo", "create", repoName, visibility, "--source=" + path, "--remote=origin"}
	if push {
		args = append(args, "--push")
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return errors.New(msg)
	}
	return nil
}

// GetRemoteWebURL returns a browser-friendly HTTPS URL for the origin remote
// of the repository at repoPath. SSH URLs (git@host:org/repo.git) are
// converted to HTTPS automatically.
func GetRemoteWebURL(repoPath string) (string, error) {
	raw, err := RunGitCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return toWebURL(strings.TrimSpace(raw)), nil
}

func toWebURL(remote string) string {
	remote = strings.TrimSuffix(remote, ".git")
	// SSH format: git@github.com:user/repo
	if strings.HasPrefix(remote, "git@") {
		remote = strings.TrimPrefix(remote, "git@")
		remote = strings.Replace(remote, ":", "/", 1)
		return "https://" + remote
	}
	return remote
}

// GetRecentCommits returns the last n commits for the repository at repoPath,
// one formatted "hash message" string per commit (most recent first).
func GetRecentCommits(repoPath string, limit int) ([]string, error) {
	output, err := RunGitCommand(repoPath, "log", fmt.Sprintf("-n%d", limit), "--oneline")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	return removeEmptyStrings(lines), nil
}

// RunGitCommand executes a git command in dir and returns its stdout as a string.
// Non-zero exit codes include git output in the returned error.
func RunGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return stdout.String(), nil
}
