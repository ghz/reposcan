package tui

import (
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
)

var reposTableKeybindings = []common.Keybinding{
	{
		Key:         "↑/↓",
		Description: "Navigate up and down (or j/k)",
		ShortDesc:   "Navigate",
	},
	{
		Key:         "tab/⇧tab",
		Description: "Cycle details: file changes / recent commits / README (⇧tab goes back)",
		ShortDesc:   "Switch Details",
	},
	{
		Key:         "+/-",
		Description: "Expand / collapse a repo to show its branches",
		ShortDesc:   "Branches",
	},
	{
		Key:         "c",
		Description: "Checkout the selected branch (branch rows only)",
		ShortDesc:   "Checkout",
	},
	{
		Key:         "g",
		Description: "Open Git actions menu",
		ShortDesc:   "Git",
	},
	{
		Key:         "w",
		Description: "Open remote URL in browser",
		ShortDesc:   "Remote",
	},
	{
		Key:         "o",
		Description: "Open in editor (VS Code by default, configurable)",
		ShortDesc:   "Open",
	},
	{
		Key:         "e",
		Description: "Open in system file manager",
		ShortDesc:   "Explorer",
	},
	{
		Key:         "f",
		Description: "Toggle favorite (pinned to top, persisted in config)",
		ShortDesc:   "Favorite",
	},
	{
		Key:         "n",
		Description: "New repo from selected folder (local or GitHub)",
		ShortDesc:   "New repo",
	},
	{
		Key:         "d",
		Description: "Delete selected local folder after typing YES",
		ShortDesc:   "Delete",
	},
	{
		Key:         "p",
		Description: "Copy repo path to clipboard",
		ShortDesc:   "Copy Path",
	},
	{
		Key:         "r",
		Description: "Refresh list",
		ShortDesc:   "Refresh",
	},
	{
		Key:         "/",
		Description: "Search by repo/branch name",
		ShortDesc:   "Search",
	},
	{
		Key:         "q",
		Description: "Quit",
		ShortDesc:   "Quit",
	},
}

// Not needed anymore. Repos table filter textfield is placed on top of footer
var reposTableFilterKeybindings = []common.Keybinding{
	{
		Key:         "<enter>",
		Description: "Apply and move cursor to repos table",
		ShortDesc:   "Apply",
	},
	{
		Key:         "<esc>",
		Description: "Hide and cancel search",
		ShortDesc:   "Cancel",
	},
}

var helpPopupKeybindings = []common.Keybinding{
	{
		Key:         "q/<esc>",
		Description: "Close Popup",
		ShortDesc:   "Close",
	},
}

var gitMenuKeybindings = []common.Keybinding{
	{
		Key:         "1",
		Description: "Quick save: git add . + commit wip + push",
		ShortDesc:   "Quick save",
	},
	{
		Key:         "2",
		Description: "Git push",
		ShortDesc:   "Push",
	},
	{
		Key:         "3",
		Description: "Git pull",
		ShortDesc:   "Pull",
	},
	{
		Key:         "4",
		Description: "Git fetch",
		ShortDesc:   "Fetch",
	},
	{
		Key:         "5",
		Description: "Open remote URL in browser",
		ShortDesc:   "Browser",
	},
	{
		Key:         "esc",
		Description: "Cancel",
		ShortDesc:   "Cancel",
	},
}

var createRepoKindKeybindings = []common.Keybinding{
	{
		Key:         "1",
		Description: "Local only",
		ShortDesc:   "Local",
	},
	{
		Key:         "2",
		Description: "GitHub private",
		ShortDesc:   "GH private",
	},
	{
		Key:         "3",
		Description: "GitHub public",
		ShortDesc:   "GH public",
	},
	{
		Key:         "esc",
		Description: "Cancel",
		ShortDesc:   "Cancel",
	},
}

var createRepoNameKeybindings = []common.Keybinding{
	{
		Key:         "enter",
		Description: "Create repo",
		ShortDesc:   "Create",
	},
	{
		Key:         "esc",
		Description: "Back to type selection",
		ShortDesc:   "Back",
	},
}

var deleteRepoKeybindings = []common.Keybinding{
	{
		Key:         "YES + enter",
		Description: "Permanently delete the repository folder",
		ShortDesc:   "Confirm",
	},
	{
		Key:         "esc",
		Description: "Cancel",
		ShortDesc:   "Cancel",
	},
}
