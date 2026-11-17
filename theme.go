package framekit

import (
	"embed"
	"encoding/hex"
	"fmt"
	"image/color"
	"strings"
)

// Fonts are embedded so the binary is self-contained and works offline.
//
//go:embed assets/fonts/GeistMono-Regular.ttf assets/fonts/GeistMono-Bold.ttf
var fontFS embed.FS

// Theme is the full visual system. The default mirrors sazardev.stream exactly:
// flat gruvbox, 1px dividers, no gradients, no shadows, no radii.
type Theme struct {
	Name   string `json:"name"`
	Brand  string `json:"brand"`
	BG     string `json:"bg"`
	Deep   string `json:"deep"`
	Soft   string `json:"soft"`
	FG     string `json:"fg"`
	Dim    string `json:"dim"`
	Faint  string `json:"faint"`
	Green  string `json:"green"`
	Yellow string `json:"yellow"`
	Aqua   string `json:"aqua"`
	Blue   string `json:"blue"`
	Red    string `json:"red"`
	Purple string `json:"purple"`
	Orange string `json:"orange"`
}

// Gruvbox returns the design system used by sazardev.stream.
func Gruvbox() Theme {
	return Theme{
		Name: "gruvbox", Brand: "SazarDev",
		BG: "#282828", Deep: "#202020", Soft: "#323232",
		FG: "#ebdbb2", Dim: "#bdae93", Faint: "#665c54",
		Green: "#b8bb26", Yellow: "#fabd2f", Aqua: "#8ec07c",
		Blue: "#83a598", Red: "#f29b7b", Purple: "#dba1a2", Orange: "#f9963e",
	}
}

// Merge overlays non-empty fields from an optional custom theme.
func (t Theme) Merge(custom *Theme) Theme {
	if custom == nil {
		return t
	}
	fields := []struct {
		dst *string
		src string
	}{
		{&t.Name, custom.Name}, {&t.Brand, custom.Brand},
		{&t.BG, custom.BG}, {&t.Deep, custom.Deep}, {&t.Soft, custom.Soft},
		{&t.FG, custom.FG}, {&t.Dim, custom.Dim}, {&t.Faint, custom.Faint},
		{&t.Green, custom.Green}, {&t.Yellow, custom.Yellow}, {&t.Aqua, custom.Aqua},
		{&t.Blue, custom.Blue}, {&t.Red, custom.Red}, {&t.Purple, custom.Purple},
		{&t.Orange, custom.Orange},
	}
	for _, f := range fields {
		if strings.TrimSpace(f.src) != "" {
			*f.dst = f.src
		}
	}
	return t
}

// Color resolves a named theme color. Unknown names fall back to FG.
func (t Theme) Color(name string) color.RGBA {
	switch strings.ToLower(name) {
	case "bg":
		return mustHex(t.BG)
	case "deep":
		return mustHex(t.Deep)
	case "soft":
		return mustHex(t.Soft)
	case "dim":
		return mustHex(t.Dim)
	case "faint":
		return mustHex(t.Faint)
	case "green":
		return mustHex(t.Green)
	case "yellow":
		return mustHex(t.Yellow)
	case "aqua":
		return mustHex(t.Aqua)
	case "blue":
		return mustHex(t.Blue)
	case "red":
		return mustHex(t.Red)
	case "purple":
		return mustHex(t.Purple)
	case "orange":
		return mustHex(t.Orange)
	default:
		return mustHex(t.FG)
	}
}

func mustHex(s string) color.RGBA {
	c, err := ParseHex(s)
	if err != nil {
		return color.RGBA{}
	}
	return c
}

// ParseHex accepts #rgb, #rgba, #rrggbb and #rrggbbaa.
func ParseHex(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 || len(s) == 4 {
		expanded := ""
		for _, ch := range s {
			expanded += string(ch) + string(ch)
		}
		s = expanded
	}
	b, err := hex.DecodeString(s)
	if err != nil || (len(b) != 3 && len(b) != 4) {
		return color.RGBA{}, fmt.Errorf("invalid color %q", s)
	}
	c := color.RGBA{R: b[0], G: b[1], B: b[2], A: 255}
	if len(b) == 4 {
		c.A = b[3]
	}
	return c, nil
}

// Mix blends a color toward the background by alpha in [0,1].
func Mix(c, bg color.RGBA, alpha float64) color.RGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	blend := func(a, b uint8) uint8 {
		return uint8(float64(a)*alpha + float64(b)*(1-alpha))
	}
	return color.RGBA{blend(c.R, bg.R), blend(c.G, bg.G), blend(c.B, bg.B), 255}
}

// FontBytes returns the embedded TTF for "regular" or "bold".
func FontBytes(weight string) ([]byte, error) {
	name := "assets/fonts/GeistMono-Regular.ttf"
	if strings.EqualFold(weight, "bold") {
		name = "assets/fonts/GeistMono-Bold.ttf"
	}
	return fontFS.ReadFile(name)
}
