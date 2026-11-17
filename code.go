package framekit

import (
	"strings"
	"unicode"
)

type token struct {
	text string
	kind string
}

var langKeywords = map[string]map[string]bool{
	"go": {"func": true, "return": true, "if": true, "else": true, "for": true, "range": true,
		"var": true, "const": true, "type": true, "struct": true, "interface": true, "package": true,
		"import": true, "go": true, "defer": true, "switch": true, "case": true, "default": true,
		"map": true, "chan": true, "select": true, "nil": true, "true": true, "false": true},
	"python": {"def": true, "return": true, "if": true, "elif": true, "else": true, "for": true,
		"in": true, "while": true, "class": true, "import": true, "from": true, "as": true,
		"with": true, "try": true, "except": true, "finally": true, "lambda": true, "yield": true,
		"None": true, "True": true, "False": true, "async": true, "await": true, "pass": true},
	"rust": {"fn": true, "let": true, "mut": true, "pub": true, "struct": true, "enum": true,
		"impl": true, "trait": true, "match": true, "if": true, "else": true, "for": true,
		"while": true, "loop": true, "return": true, "use": true, "mod": true, "crate": true,
		"self": true, "Self": true, "async": true, "await": true, "move": true, "true": true, "false": true},
	"typescript": {"const": true, "let": true, "var": true, "function": true, "return": true,
		"if": true, "else": true, "for": true, "while": true, "class": true, "interface": true,
		"type": true, "import": true, "export": true, "from": true, "default": true, "async": true,
		"await": true, "new": true, "extends": true, "implements": true, "public": true, "private": true,
		"readonly": true, "true": true, "false": true, "null": true, "undefined": true},
	"sql": {"SELECT": true, "FROM": true, "WHERE": true, "JOIN": true, "ON": true, "GROUP": true,
		"BY": true, "ORDER": true, "LIMIT": true, "INSERT": true, "INTO": true, "VALUES": true,
		"UPDATE": true, "SET": true, "DELETE": true, "CREATE": true, "TABLE": true, "AS": true,
		"AND": true, "OR": true, "NOT": true, "NULL": true, "COUNT": true, "DISTINCT": true},
	"bash": {"if": true, "then": true, "fi": true, "for": true, "do": true, "done": true,
		"while": true, "case": true, "esac": true, "function": true, "return": true, "export": true,
		"local": true, "echo": true, "cd": true, "true": true, "false": true},
	"dart": {"class": true, "void": true, "final": true, "const": true, "var": true, "if": true,
		"else": true, "for": true, "while": true, "return": true, "import": true, "extends": true,
		"with": true, "implements": true, "async": true, "await": true, "Future": true, "Stream": true,
		"true": true, "false": true, "null": true},
}

var langTypes = map[string]bool{
	"string": true, "int": true, "int64": true, "float64": true, "bool": true, "byte": true,
	"error": true, "any": true, "void": true, "String": true, "number": true, "boolean": true,
	"List": true, "Map": true, "Vec": true, "Option": true, "Result": true, "Self": true,
}

func keywordsFor(lang string) map[string]bool {
	l := strings.ToLower(lang)
	if l == "ts" || l == "js" || l == "javascript" {
		l = "typescript"
	}
	if l == "py" {
		l = "python"
	}
	if l == "rs" {
		l = "rust"
	}
	if k, ok := langKeywords[l]; ok {
		return k
	}
	return langKeywords["go"]
}

func usesHashComments(lang string) bool {
	l := strings.ToLower(lang)
	return l == "python" || l == "bash" || l == "sh" || l == "yaml" || l == "ruby"
}

// tokenize is intentionally small: it covers comments, strings, numbers and
// keywords, which is enough for readable tutorial snippets.
func tokenize(src, lang string) []token {
	kw := keywordsFor(lang)
	hash := usesHashComments(lang)
	var out []token
	i := 0
	for i < len(src) {
		rest := src[i:]
		if strings.HasPrefix(rest, "//") || (hash && strings.HasPrefix(rest, "#")) {
			out = append(out, token{rest, "comment"})
			break
		}
		if strings.HasPrefix(rest, "/*") {
			if end := strings.Index(rest, "*/"); end >= 0 {
				out = append(out, token{rest[:end+2], "comment"})
				i += end + 2
				continue
			}
			out = append(out, token{rest, "comment"})
			break
		}
		c := src[i]
		if c == '"' || c == '\'' || c == '`' {
			j := i + 1
			for j < len(src) {
				if src[j] == '\\' && j+1 < len(src) {
					j += 2
					continue
				}
				if src[j] == c {
					j++
					break
				}
				j++
			}
			out = append(out, token{src[i:j], "string"})
			i = j
			continue
		}
		if unicode.IsDigit(rune(c)) {
			j := i
			for j < len(src) && (unicode.IsDigit(rune(src[j])) || src[j] == '.' || src[j] == '_') {
				j++
			}
			out = append(out, token{src[i:j], "number"})
			i = j
			continue
		}
		if unicode.IsLetter(rune(c)) || c == '_' || c == '$' {
			j := i
			for j < len(src) && (unicode.IsLetter(rune(src[j])) || unicode.IsDigit(rune(src[j])) || src[j] == '_' || src[j] == '$') {
				j++
			}
			word := src[i:j]
			kind := "plain"
			if kw[word] || kw[strings.ToLower(word)] {
				kind = "keyword"
			} else if langTypes[word] {
				kind = "type"
			}
			out = append(out, token{word, kind})
			i = j
			continue
		}
		// plain run until the next special character
		j := i
		for j < len(src) {
			ch := src[j]
			if ch == '"' || ch == '\'' || ch == '`' || ch == '/' || ch == '#' ||
				unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) || ch == '_' {
				break
			}
			j++
		}
		if j == i {
			j++
		}
		out = append(out, token{src[i:j], "plain"})
		i = j
	}
	return out
}

func tokenColor(kind string) string {
	switch kind {
	case "keyword":
		return "purple"
	case "type":
		return "blue"
	case "string":
		return "aqua"
	case "comment":
		return "faint"
	case "number":
		return "yellow"
	default:
		return "fg"
	}
}

// drawCode renders a syntax-highlighted snippet with reveal animation.
func drawCode(p *painter, spec Spec, sc Scene, lt, fade float64) {
	if sc.Code == nil {
		return
	}
	l, S := p.lay, p.lay.Scale
	lines := strings.Split(strings.ReplaceAll(sc.Code.Content, "\t", "    "), "\n")
	x := float64(l.ContentX)
	top := float64(l.FrameY0) + 84*S
	lineH := 32 * S
	if sc.Title != "" {
		p.text(sc.Title, x, top-28*S, "h2-bold", "fg", fade)
		p.line(x, top-12*S, float64(l.FrameX1)-40*S, top-12*S, "faint", fade*0.8)
	}
	highlight := map[int]bool{}
	for _, n := range sc.Code.Highlight {
		highlight[n] = true
	}
	gutter := 0.0
	if sc.Code.LineNumbers {
		gutter = 46 * S
	}
	baseAt := sc.Code.At
	for i, line := range lines {
		at := baseAt + float64(i)*0.18
		a := fade * EasedReveal(spec, lt, at)
		if a <= 0.01 {
			continue
		}
		y := top + float64(i)*lineH
		if highlight[i+1] {
			p.ctx.SetColor(Mix(p.theme.Color("soft"), p.bg, a*0.8))
			p.ctx.DrawRectangle(x-10*S, y-24*S, float64(l.FrameX1-l.FrameX0)-70*S, lineH)
			p.ctx.Fill()
		}
		if sc.Code.LineNumbers {
			p.text(rightPad(i+1, len(lines)), x, y, "small", "faint", a*0.8)
		}
		p.ctx.SetFontFace(p.faces["code"])
		cx := x + gutter
		for _, tk := range tokenize(line, sc.Code.Language) {
			p.text(tk.text, cx, y, "code", tokenColor(tk.kind), a)
			cx += p.textWidth(tk.text, "code")
		}
	}
}

func rightPad(n, total int) string {
	s := ""
	if n >= 10 || total >= 10 {
		s = " "
	}
	return s + intStr(n) + " "
}

func intStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
