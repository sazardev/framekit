package main

import (
	"flag"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/framekit"
)

var (
	bgStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#282828")).Foreground(lipgloss.Color("#ebdbb2")).Padding(1, 2)
	accent     = lipgloss.NewStyle().Foreground(lipgloss.Color("#b8bb26")).Bold(true)
	cyanStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ec07c"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#bdae93"))
	faintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#665c54"))
	selStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#fabd2f")).Bold(true)
)

type doneMsg struct{ status string }

type tuiModel struct {
	cursor  int
	values  []string
	editing bool
	buf     string
	status  string
	busy    bool
	cfg     framekit.Config
}

var tuiFields = []string{"Template", "Format", "Quality", "Theme", "Music mood", "Title", "Output", "▶ Lint + Render", "Quit"}

func tuiCmd(args []string) {
	_ = flag.NewFlagSet("tui", flag.ExitOnError).Parse(args)
	cfg := loadCfg("")
	m := tuiModel{
		values: []string{
			"intro", "landscape", "high", "gruvbox", "ambient",
			"SazarDev — Freelance Developer Intro", "intro.mp4",
		},
		cfg:    cfg,
		status: "j/k navigate · ←/→ change · Enter edit/render · q quit",
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fatal(err)
	}
}

func (m tuiModel) Init() tea.Cmd { return nil }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.status = msg.status
		m.busy = false
		return m, nil
	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "enter":
				if m.cursor == 5 {
					m.values[5] = m.buf
				}
				if m.cursor == 6 {
					m.values[6] = m.buf
				}
				m.editing = false
				m.status = "value saved"
				return m, nil
			case "esc":
				m.editing = false
				return m, nil
			case "backspace":
				if len(m.buf) > 0 {
					m.buf = m.buf[:len(m.buf)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.buf += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			m.cursor = (m.cursor + 1) % len(tuiFields)
		case "k", "up":
			m.cursor = (m.cursor - 1 + len(tuiFields)) % len(tuiFields)
		case "h", "left":
			m.cycle(-1)
		case "l", "right":
			m.cycle(1)
		case "enter":
			switch m.cursor {
			case 5, 6:
				m.editing = true
				m.buf = m.values[m.cursor]
			case 7:
				if m.busy {
					return m, nil
				}
				m.busy = true
				m.status = "rendering…"
				cfg, vals := m.cfg, append([]string{}, m.values...)
				return m, func() tea.Msg {
					status := renderFromTUI(cfg, vals)
					return doneMsg{status: status}
				}
			case 8:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *tuiModel) cycle(dir int) {
	if m.cursor > 4 || m.busy {
		return
	}
	lists := [][]string{
		framekit.TemplateNames(m.cfg),
		{"landscape", "landscape720", "square", "portrait"},
		{"draft", "standard", "social", "high", "master"},
		framekit.ThemeNames(),
		{"ambient", "uplift", "noir", "minimal"},
	}
	list := lists[m.cursor]
	if len(list) == 0 {
		return
	}
	idx := 0
	for i, v := range list {
		if v == m.values[m.cursor] {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(list)) % len(list)
	m.values[m.cursor] = list[idx]
}

func renderFromTUI(cfg framekit.Config, vals []string) string {
	vars := map[string]string{
		"title":  vals[5],
		"format": vals[1],
		"mood":   vals[4],
		"output": vals[6],
	}
	spec, err := framekit.RenderTemplate(vals[0], vars, cfg)
	if err != nil {
		return "error: " + err.Error()
	}
	spec.Format = vals[1]
	spec.Music.Progression = vals[4]
	spec.Theme = &framekit.Theme{Name: vals[3]}
	spec.Output = vals[6]
	framekit.SanitizeSpec(&spec)
	if err := spec.Normalize(); err != nil {
		return "error: " + err.Error()
	}
	a, err := framekit.Analyze(spec, cfg, vals[2])
	if err != nil {
		return "error: " + err.Error()
	}
	findings := framekit.Lint(spec, cfg, vals[2], a)
	for _, f := range findings {
		if f.Severity == framekit.SevError {
			return fmt.Sprintf("lint error [%s]: %s", f.Code, f.Message)
		}
	}
	if err := framekit.RenderWithConfig(spec, vals[6], cfg, vals[2]); err != nil {
		return "error: " + err.Error()
	}
	return fmt.Sprintf("rendered %s · %s · %s · %.0fs", vals[6], vals[1], vals[2], a.Duration)
}

func (m tuiModel) View() string {
	var b strings.Builder
	b.WriteString(accent.Render("  FRAMEKIT  ·  VIDEO STUDIO") + "  " + faintStyle.Render(framekit.Version) + "\n\n")
	for i, field := range tuiFields {
		line := fmt.Sprintf("  %-14s", field)
		if i < len(m.values) {
			line += cyanStyle.Render(m.values[i])
			if i <= 4 {
				line += faintStyle.Render("  ‹ ›")
			}
		}
		if i == m.cursor {
			line = selStyle.Render("▸ " + strings.TrimPrefix(line, "  "))
		} else {
			line = dimStyle.Render("  " + strings.TrimPrefix(line, "  "))
		}
		if m.editing && i == m.cursor {
			line = selStyle.Render("▸ "+field) + "  " + m.buf + faintStyle.Render("█")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + faintStyle.Render("  "+m.status) + "\n")
	return bgStyle.Render(b.String())
}
