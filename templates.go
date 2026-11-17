package framekit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// BuiltinTemplates are production-ready starting points. Use {{variable}} in
// any string field; {{variable:default}} supplies a fallback.
var BuiltinTemplates = map[string]string{
	"intro": `{
  "title": "{{title:SazarDev — Freelance Developer Intro}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":true,"cursor":true,"chrome":true,"progress":true,
              "fade":0.38,"slide":10,"easing":"easeOutCubic",
              "transition":"bgfade","stagger":0.4},
  "music": {"enabled":true,"volume":0.25,"progression":"{{mood:ambient}}","tempo":80},
  "scenes": [
    {"kind":"title","name":"01 / IDENTITY","duration":3.5,
     "title":"{{name:Omar Flores}}","subtitle":"{{role:Full-Stack Freelance Developer}}","accent":"green"},
    {"kind":"terminal","name":"02 / STACK","duration":5,"command":"cat stack.txt",
     "items":[{"name":"Flutter / Dart","note":"mobile apps"},
              {"name":"React / TypeScript","note":"web apps"},
              {"name":"Go / Rust","note":"systems & CLIs"}]},
    {"kind":"stats","name":"03 / IMPACT","duration":4.5,
     "stats":[{"value":"29","label":"APPS SHIPPED"},
              {"value":"26","label":"PROJECTS"},
              {"value":"300+","label":"GITHUB STARS"}]},
    {"kind":"credits","name":"04 / CONTACT","duration":4,
     "title":"Let's build","subtitle":"something solid together.",
     "lines":[{"text":"{{cta:@salazarprograms · Fiverr}}","color":"aqua","at":1.2}]}
  ],
  "output":"{{output:intro.mp4}}"
}`,
	"reel": `{
  "title": "{{title:SazarDev — Reel}}",
  "format": "portrait",
  "effects": {"typewriter":true,"cursor":true,"chrome":true,"progress":true,
              "fade":0.32,"slide":9,"easing":"easeOutQuint","transition":"wipe",
              "scanline":true,"tracking":2.5,"stagger":0.35},
  "music": {"enabled":true,"volume":0.28,"progression":"{{mood:uplift}}","tempo":96},
  "scenes": [
    {"kind":"title","name":"01 / HOOK","duration":3,
     "title":"{{hook:Build Faster}}","subtitle":"{{sub:with the right stack}}","accent":"aqua"},
    {"kind":"terminal","name":"02 / STACK","duration":5,"command":"cat stack.txt","accent":"yellow",
     "items":[{"name":"Flutter / Dart"},{"name":"React / TypeScript"},
              {"name":"Go / Rust"},{"name":"Firebase / Postgres"}]},
    {"kind":"stats","name":"03 / PROOF","duration":4.5,"accent":"aqua",
     "stats":[{"value":"29","label":"APPS"},{"value":"26","label":"PROJECTS"},
              {"value":"300+","label":"STARS"}]},
    {"kind":"credits","name":"04 / CTA","duration":3.5,"title":"{{cta:Hire me}}",
     "subtitle":"@salazarprograms · Fiverr"}
  ],
  "output":"{{output:reel.mp4}}"
}`,
	"pitch": `{
  "title": "{{title:The Pitch}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":true,"cursor":false,"chrome":true,"progress":true,
              "fade":0.4,"slide":12,"easing":"easeOutQuint","tracking":2},
  "music": {"enabled":true,"volume":0.24,"progression":"noir","tempo":88},
  "scenes": [
    {"kind":"slide","name":"01 / PROBLEM","duration":5,"title":"{{problem:The problem}}",
     "lines":{{problemLines:[{"text":"Manual work is slow"},{"text":"Context gets lost"},{"text":"Quality varies"}]}}},
    {"kind":"slide","name":"02 / SOLUTION","duration":6,"title":"{{solution:The solution}}",
     "lines":{{solutionLines:[{"text":"One clear workflow"},{"text":"Automated checks"},{"text":"Measurable results"}]}},"accent":"aqua"},
    {"kind":"stats","name":"03 / IMPACT","duration":4,
     "stats":[{"value":"29","label":"APPS"},{"value":"26","label":"PROJECTS"},{"value":"300+","label":"STARS"}]},
    {"kind":"credits","name":"04 / CTA","duration":3.5,"title":"Let's talk",
     "subtitle":"{{cta:@salazarprograms · Fiverr}}"}
  ],
  "output":"{{output:pitch.mp4}}"
}`,
	"presentation": `{
  "title": "{{title:Presentation}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":false,"cursor":false,"chrome":true,"progress":true,
              "fade":0.35,"slide":10,"easing":"easeInOut","tracking":1.5},
  "music": {"enabled":true,"volume":0.18,"progression":"minimal","tempo":72},
  "scenes": [
    {"kind":"title","name":"00 / TITLE","duration":3,"title":"{{title:Presentation}}",
     "subtitle":"{{author:Omar Flores}}"},
    {"kind":"slide","name":"01 / AGENDA","duration":5,"title":"Agenda",
     "lines":{{agenda:[{"text":"Context"},{"text":"Approach"},{"text":"Results"},{"text":"Next steps"}]}}},
    {"kind":"slide","name":"02 / KEY POINTS","duration":6,"title":"{{section:Key points}}",
     "lines":{{points:[{"text":"Clear architecture"},{"text":"Fast iteration"},{"text":"Real metrics"}]}},"accent":"aqua"},
    {"kind":"credits","name":"03 / END","duration":3,"title":"Thank you",
     "subtitle":"{{cta:@salazarprograms}}"}
  ],
  "output":"{{output:presentation.mp4}}"
}`,
	"image-showcase": `{
  "title": "{{title:Work Showcase}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":false,"cursor":false,"chrome":true,"progress":true,
              "fade":0.5,"slide":10,"easing":"easeOutQuint"},
  "music": {"enabled":true,"volume":0.22,"progression":"ambient","tempo":76},
  "scenes": [
    {"kind":"image","name":"01 / WORK","duration":9,
     "image":{"path":"{{image:image.png}}","fit":"cover","zoom":0.08,"border":true},
     "title":"{{title:Project}}","subtitle":"{{subtitle:Built by SazarDev}}"},
    {"kind":"credits","name":"02 / CTA","duration":7,"title":"Let's build",
     "subtitle":"{{cta:@salazarprograms · Fiverr}}"}
  ],
  "output":"{{output:showcase.mp4}}"
}`,
	"quote": `{
  "title": "{{title:Quote}}",
  "format": "{{format:square}}",
  "effects": {"typewriter":true,"cursor":false,"chrome":true,"progress":true,
              "fade":0.5,"slide":14,"easing":"easeOutQuint","tracking":2.5},
  "music": {"enabled":true,"volume":0.2,"progression":"minimal","tempo":68},
  "scenes": [
    {"kind":"title","name":"01 / QUOTE","duration":10,
     "title":"{{quote:Clean code is a feature.}}",
     "subtitle":"{{author:Omar Flores}}","accent":"aqua"},
    {"kind":"credits","name":"02 / CTA","duration":6,"title":"SazarDev",
     "subtitle":"@salazarprograms · Fiverr"}
  ],
  "output":"{{output:quote.mp4}}"
}`,
}

var varPattern = regexp.MustCompile(`(?s)\{\{([a-zA-Z0-9_]+)(?::(.*?))?\}\}`)

// TemplateNames lists built-in + config-provided templates.
func TemplateNames(cfg Config) []string {
	out := make([]string, 0, len(BuiltinTemplates)+len(cfg.Templates))
	for n := range BuiltinTemplates {
		out = append(out, n)
	}
	for n := range cfg.Templates {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// RenderTemplate instantiates a template with variables.
func RenderTemplate(name string, vars map[string]string, cfg Config) (Spec, error) {
	tpl, ok := cfg.Templates[name]
	if !ok {
		tplStr, builtin := BuiltinTemplates[name]
		if !builtin {
			return Spec{}, fmt.Errorf("unknown template %q", name)
		}
		tpl = json.RawMessage(tplStr)
	}
	s := string(tpl)
	replaced := varPattern.ReplaceAllStringFunc(s, func(m string) string {
		groups := varPattern.FindStringSubmatch(m)
		key := groups[1]
		if v, ok := vars[key]; ok && v != "" {
			return v
		}
		if len(groups) > 2 {
			return strings.TrimSpace(groups[2])
		}
		return ""
	})
	var spec Spec
	if err := json.Unmarshal([]byte(replaced), &spec); err != nil {
		return Spec{}, fmt.Errorf("template %s: %w", name, err)
	}
	SanitizeSpec(&spec)
	return spec, nil
}
