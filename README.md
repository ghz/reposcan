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

### View modes — `←` / `→` to switch

| Mode | Description |
|---|---|
| **non-sync repos** | Repos with uncommitted files or unsynced commits *(default)* |
| **all repos** | Every git repository found under your roots |
| **all dirs** | All direct subdirectories of your roots — repos and plain folders |
| **non-repo dirs** | Only subdirectories that are **not** git repos |

### Keybindings

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Navigate |
| `←` / `→` or `h` / `l` | Switch view mode |
| `Tab` | Toggle details panel: **file changes** ↔ **recent commits** |
| `n` | Create a new repo from selected folder (local or GitHub) |
| `o` | Open selected repo/folder in editor (VS Code by default) |
| `g` | Open remote URL in browser |
| `f` | Toggle favorite — pinned to top, persisted in config |
| `c` | Copy path to clipboard |
| `r` | Refresh |
| `/` | Filter by repo/branch name |
| `?` | Help popup |
| `q` / `Esc` | Quit |

### Creating a repo from a folder (`n`)

In `all dirs` or `non-repo dirs` mode, press `n` on any plain folder to open a creation dialog:

```
┌──────────────────────────────────────────┐
│              Nouveau repo                │
│                                          │
│  Nom: [monprojet                       ] │
│                                          │
│  [l]  Local only                         │
│  [p]  GitHub — private                   │
│  [u]  GitHub — public                    │
│                                          │
│  [esc] Cancel                            │
└──────────────────────────────────────────┘
```

- **Local** — `git init` + `git add .` + initial commit
- **GitHub** — same, then [`gh repo create`](https://cli.github.com/) (requires the `gh` CLI installed and authenticated)

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
- [x] Details panel: file changes and recent commits
- [x] Configurable editor (`o` key)
- [x] Open remote in browser (`g` key)
- [x] Favorites pinned to top (`f` key, persisted)
- [x] Create repo from folder — local or GitHub (`n` key)
- [ ] Git push / pull / fetch from TUI
- [ ] Per-branch status view

---

## 🤝 Contributing

PRs, bug reports, and feature requests are welcome.
