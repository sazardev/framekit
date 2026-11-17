package framekit

import (
	"os"
	"sort"
)

// ThemePresets mirrors the terminal themes in sazardev's ThemeScript.astro,
// plus the default gruvbox design system. Use "name" in a spec to select one.
var ThemePresets = map[string]Theme{
	"gruvbox": Gruvbox(),
	"leader": {
		Name: "leader", BG: "#282828", Deep: "#202020", Soft: "#323232",
		FG: "#ebdbb2", Dim: "#bdae93", Faint: "#665c54",
		Green: "#b8bb26", Yellow: "#fabd2f", Aqua: "#8ec07c",
		Blue: "#83a598", Red: "#f29b7b", Purple: "#dba1a2", Orange: "#f9963e",
	},
	"catppuccin": {
		Name: "catppuccin", BG: "#1e1e2e", Deep: "#181825", Soft: "#313244",
		FG: "#cdd6f4", Dim: "#a6adc8", Faint: "#585b70",
		Green: "#a6e3a1", Yellow: "#f9e2af", Aqua: "#94e2d5",
		Blue: "#89b4fa", Red: "#f38ba8", Purple: "#cba6f7", Orange: "#fab387",
	},
	"tokyo-night": {
		Name: "tokyo-night", BG: "#1a1b26", Deep: "#15161e", Soft: "#292e42",
		FG: "#c0caf5", Dim: "#a9b1d6", Faint: "#414868",
		Green: "#9ece6a", Yellow: "#e0af68", Aqua: "#7dcfff",
		Blue: "#7aa2f7", Red: "#f7768e", Purple: "#bb9af7", Orange: "#ff9e64",
	},
	"nord": {
		Name: "nord", BG: "#2e3440", Deep: "#292e39", Soft: "#3b4252",
		FG: "#d8dee9", Dim: "#b8c0cc", Faint: "#4c566a",
		Green: "#a3be8c", Yellow: "#ebcb8b", Aqua: "#88c0d0",
		Blue: "#81a1c1", Red: "#bf616a", Purple: "#b48ead", Orange: "#d08770",
	},
	"dracula": {
		Name: "dracula", BG: "#282a36", Deep: "#21222c", Soft: "#44475a",
		FG: "#f8f8f2", Dim: "#d3d5e1", Faint: "#6272a4",
		Green: "#50fa7b", Yellow: "#f1fa8c", Aqua: "#8be9fd",
		Blue: "#bd93f9", Red: "#ff5555", Purple: "#ff79c6", Orange: "#ffb86c",
	},
	"solarized": {
		Name: "solarized", BG: "#002b36", Deep: "#073642", Soft: "#11424e",
		FG: "#eee8d5", Dim: "#93a1a1", Faint: "#586e75",
		Green: "#859900", Yellow: "#b58900", Aqua: "#2aa198",
		Blue: "#268bd2", Red: "#dc322f", Purple: "#6c71c4", Orange: "#cb4b16",
	},
	"light": {
		Name: "light", BG: "#f2e5bc", Deep: "#e9dcb3", Soft: "#e3d6ad",
		FG: "#3c3836", Dim: "#5f564f", Faint: "#a89984",
		Green: "#726d0d", Yellow: "#946110", Aqua: "#427b58",
		Blue: "#076678", Red: "#9d0006", Purple: "#823e68", Orange: "#af3a03",
	},
}

// ThemeByName returns a preset, or an empty Theme for custom hex definitions.
func ThemeByName(name string) Theme {
	if t, ok := ThemePresets[name]; ok {
		return t
	}
	return Theme{Name: name}
}

// FontBytesFrom loads a local TTF when provided, otherwise returns the
// embedded font of the requested weight.
func FontBytesFrom(path, weight string) ([]byte, error) {
	if path == "" {
		return FontBytes(weight)
	}
	return os.ReadFile(path)
}

// ThemeNames returns sorted preset names for CLI help.
func ThemeNames() []string {
	out := make([]string, 0, len(ThemePresets))
	for n := range ThemePresets {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
