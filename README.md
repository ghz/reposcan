# RepoScan

`reposcan` is a command-line tool written in Go that scans your filesystem for Git repositories and reports their status.  
It helps you quickly find repositories with uncommitted files, unpushed commits, or unpulled changes — and act on them directly from an interactive TUI.

🖼 Demo

https://github.com/user-attachments/assets/1c8370c6-3b94-4490-bc96-fc179ef14f1d

---

## ✨ Use cases

- **Daily sync check**: See which repos have uncommitted work or unsynced commits before switching machines.
- **Context switch**: Know what's dirty before you leave for the day.
- **Housekeeping**: Find folders that aren't git repos yet and initialize them in one keystroke.
- **Automation**: Export JSON reports to integrate with dashboards or other tools.

---

## 📦 Installation

### Install script (recommended)

**Linux / macOS**

```sh
curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | sh
```

Supports **linux/amd64**, **darwin/amd64**, **darwin/arm64**.

**Windows**

```powershell
irm https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.ps1 | iex
```

Supports **windows/amd64**.

#### Migrating from `go install`

If you previously installed reposcan via `go install`, the old binary in `$GOPATH/bin` may take precedence. Remove it first:

```sh
rm "$(which reposcan)"
```

Then reinstall with the script above.

### From source

```sh
git clone https://github.com/mabd-dev/reposcan.git
cd reposcan

# Linux / macOS
go build -o reposcan .

# Windows
go build -o reposcan.exe .
```

---

## 🚀 Usage

```sh
# Scan your home directory
reposcan

# Custom root
reposcan -r ~/Code

# Multiple roots
reposcan -r ~/Code -r ~/work
```

### Flags

```
-r, --root stringArray          Root directory to scan (repeatable). Defaults to $HOME.
-d, --dirIgnore stringArray     Glob patterns to ignore during scan (repeatable)
-f, --filter string             Repository filter: all|dirty|uncommitted|unpushed|unpulled (default "dirty")
-o, --output string             Output format: json|interactive|none (default "interactive")
    --editor string             CLI command used to open repos/folders (e.g. code, zed, idea)
    --json-output-path string   Write scan report JSON files to this directory (optional)
-w, --max-workers int           Number of concurrent git checks (default 8)
    --debug                     Enable debug logging
-h, --help                      Help
```

---

## 🖥 Interactive TUI

Launch the TUI with the default `interactive` output:

```sh
reposcan
```

### View modes — `Tab` / `Shift+Tab` to switch

| Mode | Description |
|---|---|
| **non-sync repos** | Repos with uncommitted files or unsynced commits *(default)* |
| **all repos** | Every git repository found under your roots |
| **all dirs** | All direct subdirectories of your roots — repos and plain folders |
| **non-repo dirs** | Only subdirectories that are **not** git repos |

### Keybindings

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Switch view mode |
| `↑` / `↓` or `j` / `k` | Navigate |
| `←` / `→` or `h` / `l` | Switch details tab: **file changes** / **recent commits** / **README** |
| `+` / `-` | Expand / collapse a repo to show its branches inline (ahead/behind per branch) |
| `c` | Checkout the selected branch (only on expanded branch rows) |
| `g` | Open the **Git actions** menu (quick save, commit, push, pull, fetch, browser) |
| `w` | Open remote URL in browser |
| `o` | Open selected repo/folder in editor (VS Code by default, configurable) |
| `e` | Open selected repo/folder in the system file manager |
| `f` | Toggle favorite — pinned to top, persisted in config |
| `n` | Create a new repo from selected folder (local or GitHub) |
| `d` | Delete the selected local folder (typed `YES` confirmation) |
| `p` | Copy path to clipboard |
| `r` | Refresh |
| `/` | Search by repo/branch name |
| `?` | Help popup |
| `q` / `Esc` | Quit |

### Git actions menu (`g`)

Press `g` on a git repository to open a menu of git operations:

```
┌──────────────────────────────────────────┐
│               Git                        │
│                                          │
│  Project  reposcan                       │
│  Branch   main                           │
│  ────────────────────────────────────    │
│  Action:                                 │
│                                          │
│  [1]  Quick save                          │
│  [2]  Commit…                             │
│  [3]  Push                                │
│  [4]  Pull                                │
│  [5]  Fetch                               │
│  [6]  Open remote                         │
│                                          │
│  [esc] Cancel                            │
└──────────────────────────────────────────┘
```

- The menu header shows the **project** and the **checked-out branch** the actions
  will run on — these always target the repo's current branch, even if a different
  branch row is highlighted.
- **Quick save** runs `git add .`, commits a `wip` snapshot, then pushes — one
  keystroke, no prompt.
- **Commit…** runs `git add .`, then opens a popup to type a commit message and
  commits with it. It does **not** push, leaving the commit for you to review or
  amend. An empty message defaults to `wip`.
- The repo's state in the table refreshes once the operation finishes.

### Creating a repo from a folder (`n`)

In `all dirs` or `non-repo dirs` mode, press `n` on any plain folder to open a
two-step creation dialog — first pick the repo type, then enter a name:

```
┌──────────────────────────────────────────┐
│                New repo                  │
│                                          │
│  Repo type:                               │
│                                          │
│  [1]  Local only                          │
│  [2]  GitHub private                      │
│  [3]  GitHub public                       │
│                                          │
│  [esc] Cancel                            │
└──────────────────────────────────────────┘
```

- **Local** — `git init` + `git add .` + initial commit
- **GitHub** — same, then [`gh repo create`](https://cli.github.com/) (requires the `gh` CLI installed and authenticated)

### Deleting a folder (`d`)

Press `d` on a selected folder to permanently delete it from disk. The dialog
requires you to type `YES` and press `enter` to confirm — `esc` cancels.

---

## ⚙️ Configuration

Default location:

```
# Linux / macOS
~/.config/reposcan/config.toml

# Windows
%USERPROFILE%\.config\reposcan\config.toml
```

Example:

```toml
version = 1
debug = false

# Directories to scan
roots = ["~/Code", "~/work"]

# Only show: dirty | all | uncommitted | unpushed | unpulled
only = "dirty"

# Editor command for the `o` keybinding
editor = "code"

# Pinned repos (managed by `f` in the TUI)
favorites = [
  "/home/user/Code/myproject",
]

# Skip these directories (glob patterns)
dirIgnore = [
  "**/node_modules/**",
  "**/.cache/**",
]

[output]
type = "interactive"
jsonPath = ""
```

> CLI flags override config file values.

See [sample/config.toml](sample/config.toml) for the full reference.

### Config lookup order

1. Built-in defaults
2. `~/.config/reposcan/config.toml`
3. CLI flags

Each step overrides the one before it.

---

## 🛣 Roadmap

- [x] Scan filesystem for git repos
- [x] Detect uncommitted files, unpushed and unpulled commits
- [x] Output formats: interactive TUI, JSON, none
- [x] User-configurable `config.toml`
- [x] Export report to JSON file
- [x] `dirIgnore` glob patterns
- [x] Worker pool for concurrent git checks
- [x] Windows support
- [x] Git worktree support
- [x] 4-mode tab view (non-sync / all repos / all dirs / non-repo dirs)
- [x] Details panel: file changes, recent commits and README tabs
- [x] Expandable repo tree — `+`/`-` to list each repo's branches inline
- [x] Configurable editor (`o` key)
- [x] Open remote in browser (`g` key)
- [x] Favorites pinned to top (`f` key, persisted)
- [x] Create repo from folder — local or GitHub (`n` key)
- [x] Git actions menu — quick save, commit, push, pull, fetch (`g` key)
- [x] Delete folder from disk with typed confirmation (`d` key)
- [x] Open in system file manager (`e` key)

---

## 🤝 Contributing

PRs, bug reports, and feature requests are welcome.
