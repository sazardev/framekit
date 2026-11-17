package framekit

import (
	"strings"
)

// mathItem is one renderable atom of an equation.
type mathItem struct {
	kind        string // normal | sup | sub | frac | sqrt
	text        string
	top, bottom string
}

var mathSymbols = map[string]string{
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ", "epsilon": "ε",
	"theta": "θ", "lambda": "λ", "mu": "μ", "sigma": "σ", "pi": "π",
	"phi": "φ", "omega": "ω", "rho": "ρ", "tau": "τ",
	"Delta": "Δ", "Sigma": "Σ", "Omega": "Ω", "Gamma": "Γ", "Theta": "Θ",
	"sum": "∑", "int": "∫", "partial": "∂", "infty": "∞", "nabla": "∇",
	"to": "→", "rightarrow": "→", "leftarrow": "←", "Rightarrow": "⇒",
	"approx": "≈", "neq": "≠", "leq": "≤", "geq": "≥", "times": "×",
	"cdot": "·", "div": "÷", "pm": "±", "in": "∈", "notin": "∉",
	"forall": "∀", "exists": "∃", "emptyset": "∅", "therefore": "∴",
	"log": "log", "ln": "ln", "sin": "sin", "cos": "cos", "tan": "tan",
	"lim": "lim", "exp": "exp", "max": "max", "min": "min", "mod": "mod",
}

// parseMath understands: \frac{a}{b}, \sqrt{x}, ^{...}, _{...}, ^2, _i,
// \alpha-style commands and unicode already present in the string.
func parseMath(src string) []mathItem {
	var out []mathItem
	i := 0
	for i < len(src) {
		c := src[i]
		if c == '\\' {
			cmd, next := readCommand(src, i+1)
			switch cmd {
			case "frac":
				num, j := readGroup(src, next)
				den, k := readGroup(src, j)
				out = append(out, mathItem{kind: "frac", top: num, bottom: den})
				i = k
				continue
			case "sqrt":
				inner, j := readGroup(src, next)
				out = append(out, mathItem{kind: "sqrt", text: inner})
				i = j
				continue
			default:
				if sym, ok := mathSymbols[cmd]; ok {
					out = append(out, mathItem{kind: "normal", text: sym})
				} else if cmd != "" {
					out = append(out, mathItem{kind: "normal", text: cmd})
				}
				i = next
				continue
			}
		}
		if c == '^' || c == '_' {
			kind := "sup"
			if c == '_' {
				kind = "sub"
			}
			if i+1 < len(src) && src[i+1] == '{' {
				inner, j := readGroup(src, i+1)
				out = append(out, mathItem{kind: kind, text: flattenGroup(inner)})
				i = j
				continue
			}
			if i+1 < len(src) {
				out = append(out, mathItem{kind: kind, text: flattenGroup(string(src[i+1]))})
				i += 2
				continue
			}
		}
		if c == '{' || c == '}' {
			i++
			continue
		}
		// plain run
		j := i
		for j < len(src) {
			ch := src[j]
			if ch == '\\' || ch == '^' || ch == '_' || ch == '{' || ch == '}' {
				break
			}
			j++
		}
		out = append(out, mathItem{kind: "normal", text: src[i:j]})
		i = j
	}
	return out
}

// flattenGroup expands commands and unicode sub/superscripts inside a group,
// so nested expressions like {\log_2 2} render as "log₂ 2".
func flattenGroup(s string) string {
	for cmd, sym := range mathSymbols {
		s = strings.ReplaceAll(s, "\\"+cmd, sym)
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if (s[i] == '_' || s[i] == '^') && i+1 < len(s) {
			sub := s[i] == '_'
			if s[i+1] == '{' {
				inner, j := readGroup(s, i+1)
				b.WriteString(toSubSuper(inner, sub))
				i = j
				continue
			}
			b.WriteString(toSubSuper(string(s[i+1]), sub))
			i += 2
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func toSubSuper(s string, sub bool) string {
	subDigits := []string{"₀", "₁", "₂", "₃", "₄", "₅", "₆", "₇", "₈", "₉"}
	supDigits := []string{"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"}
	out := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			if sub {
				out += subDigits[r-'0']
			} else {
				out += supDigits[r-'0']
			}
			continue
		}
		out += string(r)
	}
	return out
}

func readCommand(src string, i int) (string, int) {
	j := i
	for j < len(src) && ((src[j] >= 'a' && src[j] <= 'z') || (src[j] >= 'A' && src[j] <= 'Z')) {
		j++
	}
	if j == i && j < len(src) {
		// symbol command like \% -> literal %
		return string(src[j]), j + 1
	}
	return src[i:j], j
}

func readGroup(src string, i int) (string, int) {
	if i >= len(src) || src[i] != '{' {
		if i < len(src) {
			return string(src[i]), i + 1
		}
		return "", i
	}
	depth := 0
	j := i
	for j < len(src) {
		if src[j] == '{' {
			depth++
		} else if src[j] == '}' {
			depth--
			if depth == 0 {
				return src[i+1 : j], j + 1
			}
		}
		j++
	}
	return src[i+1:], j
}

// mathWidth measures a rendered equation using the painter's current face.
func mathWidth(p *painter, items []mathItem) float64 {
	w := 0.0
	for _, it := range items {
		switch it.kind {
		case "frac":
			a := p.textWidth(it.top, "normal")
			b := p.textWidth(it.bottom, "normal")
			w += maxFloat(a, b) + 14*p.lay.Scale
		case "sqrt":
			w += p.textWidth(it.text, "normal") + 26*p.lay.Scale
		default:
			w += p.textWidth(it.text, "normal")
		}
	}
	return w
}

// drawMath renders items centered at (cx, baseline).
func drawMath(p *painter, items []mathItem, cx, baseline float64, alpha float64) {
	drawMathAt(p, items, cx-mathWidth(p, items)/2, baseline, alpha)
}

// drawMathAt renders items starting at a left x coordinate.
func drawMathAt(p *painter, items []mathItem, x, baseline float64, alpha float64) {
	S := p.lay.Scale
	for _, it := range items {
		switch it.kind {
		case "frac":
			tw := maxFloat(p.textWidth(it.top, "normal"), p.textWidth(it.bottom, "normal"))
			p.textAnchored(it.top, x+tw/2, baseline-16*S, "normal", "fg", alpha)
			p.textAnchored(it.bottom, x+tw/2, baseline+16*S, "normal", "fg", alpha)
			p.ctx.SetColor(Mix(p.theme.Color("green"), p.bg, alpha))
			p.ctx.SetLineWidth(1)
			p.ctx.DrawLine(x, baseline, x+tw, baseline)
			p.ctx.Stroke()
			x += tw + 14*S
		case "sqrt":
			p.text("√", x, baseline+8*S, "h2", "aqua", alpha)
			iw := p.textWidth(it.text, "normal")
			p.text(it.text, x+20*S, baseline, "normal", "fg", alpha)
			p.ctx.SetColor(Mix(p.theme.Color("aqua"), p.bg, alpha))
			p.ctx.SetLineWidth(1)
			p.ctx.DrawLine(x+18*S, baseline-22*S, x+20*S+iw, baseline-22*S)
			p.ctx.Stroke()
			x += iw + 26*S
		case "sup":
			p.text(it.text, x, baseline-14*S, "small", "yellow", alpha)
			x += p.textWidth(it.text, "small")
		case "sub":
			p.text(it.text, x, baseline+12*S, "small", "blue", alpha)
			x += p.textWidth(it.text, "small")
		default:
			p.text(it.text, x, baseline, "normal", "fg", alpha)
			x += p.textWidth(it.text, "normal")
		}
	}
}

// mathValid checks balanced braces and roughly known commands (soft check).
func mathValid(expr string) bool {
	depth := 0
	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// mathHasUnknownCommand reports commands not in the symbol table or special list.
func mathHasUnknownCommand(expr string) []string {
	special := map[string]bool{"frac": true, "sqrt": true}
	unknown := []string{}
	for i := 0; i < len(expr); i++ {
		if expr[i] != '\\' {
			continue
		}
		cmd, next := readCommand(expr, i+1)
		_, known := mathSymbols[cmd]
		if !special[cmd] && !known {
			unknown = append(unknown, cmd)
		}
		i = next - 1
	}
	return unknown
}

func mathPlain(expr string) string {
	items := parseMath(expr)
	parts := []string{}
	for _, it := range items {
		switch it.kind {
		case "frac":
			parts = append(parts, it.top+"/"+it.bottom)
		case "sqrt":
			parts = append(parts, "√"+it.text)
		default:
			parts = append(parts, it.text)
		}
	}
	return strings.Join(parts, "")
}
