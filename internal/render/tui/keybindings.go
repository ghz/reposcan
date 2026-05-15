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
		Key:         "S",
		Description: "Quick save: git add . + commit wip + push",
		ShortDesc:   "Quick save",
	},
	{
		Key:         "P",
		Description: "Git push",
		ShortDesc:   "Push",
	},
	{
		Key:         "p",
		Description: "Git pull",
		ShortDesc:   "Pull",
	},
	{
		Key:         "F",
		Description: "Git fetch",
		ShortDesc:   "Fetch",
	},
	{
		Key:         "d",
		Description: "Toggle details: file changes / recent commits",
		ShortDesc:   "Details",
	},
	{
		Key:         "o",
		Description: "Open in editor (VS Code by default, configurable)",
		ShortDesc:   "Open",
	},
	{
		Key:         "g",
		Description: "Open remote URL in browser",
		ShortDesc:   "Browser",
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
		Key:         "c",
		Description: "Copy repo path to clipboard",
		ShortDesc:   "Copy Path",
	},
	{
		Key:         "r",
		Description: "Refresh list",
		ShortDesc:   "Refresh list",
	},
	{
		Key:         "/",
		Description: "Filter by repo/branch name",
		ShortDesc:   "Filter",
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
		Description: "Hide and cancel filter",
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

var createRepoPopupKeybindings = []common.Keybinding{
	{
		Key:         "l",
		Description: "Local uniquement",
		ShortDesc:   "Local",
	},
	{
		Key:         "p",
		Description: "GitHub — privé",
		ShortDesc:   "GH privé",
	},
	{
		Key:         "u",
		Description: "GitHub — public",
		ShortDesc:   "GH public",
	},
	{
		Key:         "esc",
		Description: "Annuler",
		ShortDesc:   "Annuler",
	},
}
