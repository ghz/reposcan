package config

import "os"

// Config holds all runtime options used by reposcan.
// Values may come from a config file and/or be overridden by CLI flags.
type Config struct {
	Roots     []string   `toml:"roots,omitempty"`
	DirIgnore []string   `toml:"dirignore,omitempty"`
	Only      OnlyFilter `toml:"only,omitempty"`

	Output Output `toml:"output"`

	// Editor is the CLI command used to open repos/folders (e.g. "code", "zed", "idea").
	Editor string `toml:"editor,omitempty"`

	// Favorites is a list of repo paths pinned to the top of the list.
	Favorites []string `toml:"favorites,omitempty"`

	// Max git checker workers
	MaxWorkers int `toml:"maxWorkers"`

	// Debug if true, enable logging to a file in [DefaultLogFileDir]
	Debug bool `toml:"debug"`

	Version int `toml:"version"`

	// ConfigFilePath is the absolute path to the loaded config file.
	// Not serialized — set at runtime.
	ConfigFilePath string `toml:"-"`
}

// Defaults returns a Config populated with sensible defaults suitable for
// typical local development machines.
func Defaults() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	var roots []string = nil
	roots = []string{home}

	defaultDirIgnore := []string{
		// --- Package managers / deps ---
		"**/node_modules/**",
		"**/vendor/**",
		"**/.venv/**",
		"**/venv/**",
		"**/.m2/**",
		"**/.gradle/**",
		"**/.cargo/**",
		"**/.gradle/**",
		"**/.kotlin/**",
		"**/.java/**",
		"**/.cargo/**",
		"**/.zen/**",
		"**/.bun/**",
		"**/.codex/**",
		"**/.android/**",
		"**/.config/Google/**",
		"**/.config/JetBrains/**",
		"**/target/**",

		// --- Build / dist ---
		"**/build/**",
		"**/dist/**",
		"**/.next/**",
		"**/.nuxt/**",

		// --- Cache & temp ---
		"**/.cache/**",
		"**/.local/**",
		"**/.pytest_cache/**",

		// --- IDE / tooling ---
		"**/.idea/**",
		"**/.vscode/**",
		"**/.terraform/**",
		"**/.docker/**",

		// --- OS metadata ---
		"**/.DS_Store", // macOS
		"**/Thumbs.db", // Windows

		// --- Linux system dirs ---
		"/proc/**",
		"/sys/**",
		"/dev/**",
		"/run/**",
		"/tmp/**",
		"/var/log/**",
		"/var/tmp/**",

		// --- macOS system dirs ---
		"/System/**",
		"/Library/**",
		"~/Library/**",

		// --- Windows system dirs ---
		"C:/Windows/**",
		"C:/Program Files/**",
		"C:/Program Files (x86)/**",
	}

	newOutput := Output{
		Type:     OutputInteractive,
		JSONPath: "",
	}

	return Config{
		Roots:      roots,
		DirIgnore:  defaultDirIgnore,
		Only:       OnlyDirty,
		Output:     newOutput,
		Editor:     "code",
		MaxWorkers: 8,
		Debug:      false,
		Version:    1,
	}
}
