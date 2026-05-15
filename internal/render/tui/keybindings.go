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
		Key:         "tab",
		Description: "Toggle details: file changes / recent commits",
		ShortDesc:   "Details",
	},
	{
		Key:         "g",
		Description: "Open Git actions menu",
		ShortDesc:   "Git",
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
		Key:         "c",
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
		Description: "Local uniquement",
		ShortDesc:   "Local",
	},
	{
		Key:         "2",
		Description: "GitHub — privé",
		ShortDesc:   "GH privé",
	},
	{
		Key:         "3",
		Description: "GitHub — public",
		ShortDesc:   "GH public",
	},
	{
		Key:         "esc",
		Description: "Annuler",
		ShortDesc:   "Annuler",
	},
}

var createRepoNameKeybindings = []common.Keybinding{
	{
		Key:         "enter",
		Description: "Créer le repo",
		ShortDesc:   "Créer",
	},
	{
		Key:         "esc",
		Description: "Retour au choix du type",
		ShortDesc:   "Retour",
	},
}
