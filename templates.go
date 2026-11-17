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
  "effects": {"typewriter":true,"cursor":true,"chrome":false,"progress":false,
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
  "effects": {"typewriter":true,"cursor":true,"chrome":false,"progress":false,
              "fade":0.32,"slide":9,"easing":"easeOutQuint","transition":"none",
              "scanline":false,"tracking":2.5,"stagger":0.35},
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
  "effects": {"typewriter":true,"cursor":false,"chrome":false,"progress":false,
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
  "effects": {"typewriter":false,"cursor":false,"chrome":false,"progress":false,
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
  "effects": {"typewriter":false,"cursor":false,"chrome":false,"progress":false,
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
	"explainer": `{
  "title": "{{title:Big O Notation, Visually}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":true,"cursor":true,"chrome":false,"progress":false,
              "fade":0.4,"slide":12,"easing":"easeOutQuint","tracking":1.5},
  "music": {"enabled":true,"volume":0.2,"progression":"minimal","tempo":72},
  "scenes": [
    {"kind":"title","name":"01 / HOOK","duration":3.5,
     "title":"{{title:Big O Notation, Visually}}",
     "subtitle":"{{subtitle:why growth rate matters}}","accent":"aqua"},
    {"kind":"formula","name":"02 / EQUATION","duration":9,
     "title":"{{formulaTitle:The complexity}}",
     "formula":{"expr":"T(n) = O(n \\log n)","animate":true,
       "caption":"n items, split in half, merged back together",
       "steps":[
         {"expr":"T(n) = 2T(n/2) + O(n)","note":"recurrence"},
         {"expr":"n^{\\log_2 2} = n","note":"master theorem"},
         {"expr":"T(n) = O(n \\log n)","note":"result"}]}},
    {"kind":"code","name":"03 / CODE","duration":7,
     "title":"{{codeTitle:Where it comes from}}",
     "code":{"language":"go","lineNumbers":true,"highlight":[2,3],
       "content":"func mergeSort(a []int) []int {\n  mid := len(a) / 2\n  left := mergeSort(a[:mid])\n  right := mergeSort(a[mid:])\n  return merge(left, right)\n}"}},
    {"kind":"diagram","name":"04 / DIAGRAM","duration":6,
     "title":"{{diagramTitle:Split, then merge}}",
     "diagram":{"layout":"row",
       "nodes":[{"id":"in","label":"input"},{"id":"split","label":"split","accent":"yellow"},{"id":"merge","label":"merge","accent":"aqua"}],
       "edges":[{"from":"in","to":"split"},{"from":"split","to":"merge","label":"log n levels"}]}},
    {"kind":"draw","name":"05 / WHITEBOARD","duration":7,
     "title":"{{drawTitle:See the curve}}",
     "drawing":{"strokes":[
       {"kind":"line","points":[{"x":0.12,"y":0.8},{"x":0.12,"y":0.15}],"at":0.2,"duration":1,"label":"time"},
       {"kind":"line","points":[{"x":0.12,"y":0.8},{"x":0.9,"y":0.8}],"at":1.2,"duration":1,"label":"n"},
       {"kind":"path","points":[{"x":0.14,"y":0.78},{"x":0.35,"y":0.68},{"x":0.55,"y":0.52},{"x":0.75,"y":0.34},{"x":0.88,"y":0.2}],"at":2.4,"duration":2.2,"label":"n log n","color":"aqua"}]}},
    {"kind":"credits","name":"06 / END","duration":4,
     "title":"Key takeaway","subtitle":"{{cta:Follow for more}}"}
  ],
  "output":"{{output:explainer.mp4}}"
}`,
	"tutorial": `{
  "title": "{{title:Understanding Channels in Go}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":true,"cursor":true,"chrome":false,"progress":false,
              "fade":0.36,"slide":10,"easing":"easeOutCubic","stagger":0.35},
  "music": {"enabled":true,"volume":0.2,"progression":"minimal","tempo":76},
  "scenes": [
    {"kind":"title","name":"01 / HOOK","duration":3.5,
     "title":"{{title:Understanding Channels in Go}}",
     "subtitle":"{{subtitle:concurrency, explained visually}}","accent":"aqua"},
    {"kind":"code","name":"02 / CODE","duration":7,
     "title":"{{codeTitle:The pattern}}",
     "code":{"language":"{{language:go}}","lineNumbers":true,"highlight":[4,5],
       "content":"package main\n\nfunc main() {\n  ch := make(chan int)\n  go func() { ch <- 42 }()\n  fmt.Println(<-ch)\n}"}},
    {"kind":"diagram","name":"03 / DIAGRAM","duration":7,
     "title":"{{diagramTitle:How data flows}}",
     "diagram":{"layout":"row",
       "nodes":[{"id":"a","label":"producer"},{"id":"b","label":"channel","accent":"yellow"},{"id":"c","label":"consumer"}],
       "edges":[{"from":"a","to":"b","label":"send"},{"from":"b","to":"c","label":"recv"}]}},
    {"kind":"draw","name":"04 / WHITEBOARD","duration":8,
     "title":"{{drawTitle:Putting it together}}",
     "drawing":{"strokes":[
       {"kind":"rect","x":0.08,"y":0.25,"w":0.24,"h":0.5,"at":0.3,"duration":1.2,"label":"goroutine"},
       {"kind":"arrow","points":[{"x":0.34,"y":0.5},{"x":0.58,"y":0.5}],"at":1.8,"duration":1,"label":"chan int"},
       {"kind":"circle","x":0.78,"y":0.5,"w":0.14,"h":0.26,"at":3,"duration":1.2,"label":"main"}]}},
    {"kind":"credits","name":"05 / RECAP","duration":4,
     "title":"Key takeaways","subtitle":"{{cta:Follow for more}}"}
  ],
  "output":"{{output:tutorial.mp4}}"
}`,
	"theory": `{
  "title": "{{title:Software Architecture Basics}}",
  "format": "{{format:landscape}}",
  "effects": {"typewriter":false,"cursor":false,"chrome":false,"progress":false,
              "fade":0.45,"slide":12,"easing":"easeInOut","tracking":1.5},
  "music": {"enabled":true,"volume":0.18,"progression":"ambient","tempo":68},
  "scenes": [
    {"kind":"title","name":"00 / TITLE","duration":3,
     "title":"{{title:Software Architecture Basics}}","subtitle":"{{subtitle:theory, visually}}"},
    {"kind":"slide","name":"01 / CONCEPT","duration":5,"title":"{{section:The problem}}",
     "lines":[{"text":"Growing systems need boundaries"},{"text":"Dependencies must be explicit"},{"text":"Change should be local"}]},
    {"kind":"diagram","name":"02 / MODEL","duration":7,"title":"{{diagramTitle:Layered architecture}}",
     "diagram":{"layout":"row",
       "nodes":[{"id":"ui","label":"UI","accent":"aqua"},{"id":"svc","label":"Service"},{"id":"db","label":"Storage","accent":"yellow"}],
       "edges":[{"from":"ui","to":"svc"},{"from":"svc","to":"db"}]}},
    {"kind":"draw","name":"03 / WHITEBOARD","duration":7,"title":"{{drawTitle:Draw it yourself}}",
     "drawing":{"strokes":[
       {"kind":"rect","x":0.1,"y":0.2,"w":0.3,"h":0.6,"at":0.3,"duration":1.5,"label":"domain"},
       {"kind":"arrow","points":[{"x":0.42,"y":0.5},{"x":0.62,"y":0.5}],"at":2,"duration":1},
       {"kind":"rect","x":0.64,"y":0.2,"w":0.26,"h":0.6,"at":3,"duration":1.5,"label":"infra"}]}},
    {"kind":"credits","name":"04 / END","duration":3.5,"title":"Summary",
     "subtitle":"{{cta:@salazarprograms · Fiverr}}"}
  ],
  "output":"{{output:theory.mp4}}"
}`,
	"quote": `{
  "title": "{{title:Quote}}",
  "format": "{{format:square}}",
  "effects": {"typewriter":true,"cursor":false,"chrome":false,"progress":false,
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
