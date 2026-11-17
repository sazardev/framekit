package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sazardev/framekit"
)

func loadCfg(path string) framekit.Config {
	cfg, err := framekit.LoadConfig(path)
	if err != nil {
		fatal(err)
	}
	return cfg
}

func lintCmd(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	p := fs.String("spec", "", "spec path")
	configPath := fs.String("config", "", "config path")
	quality := fs.String("quality", "", "quality profile")
	jsonOut := fs.Bool("json", false, "JSON report")
	_ = fs.Parse(args)
	spec := loadSpec(*p)
	cfg := loadCfg(*configPath)
	framekit.SanitizeSpec(&spec)
	if err := spec.Normalize(); err != nil {
		fatal(err)
	}
	a, err := framekit.Analyze(spec, cfg, *quality)
	if err != nil {
		fatal(err)
	}
	findings := append(framekit.Lint(spec, cfg, *quality, a), framekit.SecurityAudit(spec, cfg)...)
	if *jsonOut {
		b, _ := json.MarshalIndent(map[string]any{"analysis": a, "findings": findings}, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("Lint · %s · %s · %s · %.0fs · ~%.2f MB\n", spec.Title, spec.Format, a.Quality, a.Duration, a.EstimatedMB)
		if len(findings) == 0 {
			fmt.Println("  no findings — spec is clean")
		}
		for _, f := range findings {
			scene := ""
			if f.Scene > 0 {
				scene = fmt.Sprintf(" [scene %d]", f.Scene)
			}
			fmt.Printf("  %-5s %-18s %s%s\n", strings.ToUpper(string(f.Severity)), f.Code, f.Message, scene)
		}
	}
	for _, f := range findings {
		if f.Severity == framekit.SevError {
			os.Exit(2)
		}
	}
}

func analyzeCmd(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	p := fs.String("spec", "", "spec path")
	configPath := fs.String("config", "", "config path")
	quality := fs.String("quality", "", "quality profile")
	_ = fs.Parse(args)
	spec := loadSpec(*p)
	cfg := loadCfg(*configPath)
	a, err := framekit.Analyze(spec, cfg, *quality)
	if err != nil {
		fatal(err)
	}
	b, _ := json.MarshalIndent(a, "", "  ")
	fmt.Println(string(b))
}

func templatesCmd(args []string) {
	fs := flag.NewFlagSet("templates", flag.ExitOnError)
	configPath := fs.String("config", "", "config path")
	_ = fs.Parse(args)
	cfg := loadCfg(*configPath)
	for _, n := range framekit.TemplateNames(cfg) {
		fmt.Println(n)
	}
}

func useCmd(args []string) {
	fs := flag.NewFlagSet("use", flag.ExitOnError)
	name := fs.String("template", "", "template name")
	varsPath := fs.String("vars", "", "JSON variables file")
	out := fs.String("out", "spec.json", "output spec")
	configPath := fs.String("config", "", "config path")
	_ = fs.Parse(args)
	if *name == "" {
		fatal(fmt.Errorf("--template is required"))
	}
	vars := map[string]string{}
	if *varsPath != "" {
		b, err := os.ReadFile(*varsPath)
		if err != nil {
			fatal(err)
		}
		if err := json.Unmarshal(b, &vars); err != nil {
			fatal(err)
		}
	}
	cfg := loadCfg(*configPath)
	spec, err := framekit.RenderTemplate(*name, vars, cfg)
	if err != nil {
		fatal(err)
	}
	b, _ := json.MarshalIndent(spec, "", "  ")
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Println("wrote", *out)
}

func generateCmd(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	p := fs.String("spec", "", "base spec")
	count := fs.Int("count", 3, "number of variants")
	seed := fs.Int64("seed", 42, "deterministic seed")
	dir := fs.String("dir", "variants", "output directory")
	render := fs.Bool("render", false, "render each variant")
	quality := fs.String("quality", "", "quality profile")
	_ = fs.Parse(args)
	base := loadSpec(*p)
	cfg := loadCfg("")
	variants, err := framekit.Generate(base, *count, *seed)
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		fatal(err)
	}
	for i, v := range variants {
		b, _ := json.MarshalIndent(v, "", "  ")
		name := filepath.Join(*dir, fmt.Sprintf("variant-%02d.json", i+1))
		if err := os.WriteFile(name, append(b, '\n'), 0o644); err != nil {
			fatal(err)
		}
		fmt.Println("wrote", name)
		if *render {
			out := strings.TrimSuffix(name, ".json") + ".mp4"
			if err := framekit.RenderWithConfig(v, out, cfg, *quality); err != nil {
				fatal(err)
			}
			fmt.Println("rendered", out)
		}
	}
}

func qualityCmd(args []string) {
	_ = flag.NewFlagSet("quality", flag.ExitOnError).Parse(args)
	cfg := loadCfg("")
	names := make([]string, 0, len(cfg.Quality))
	for n := range cfg.Quality {
		names = append(names, n)
	}
	// stable-ish order
	for _, n := range []string{"draft", "standard", "high", "social", "master"} {
		q := cfg.Quality[n]
		fmt.Printf("%-9s scale %.2f · %dfps · crf %d · %s · max %s · audio %s\n",
			n, q.Scale, q.FPS, q.CRF, q.Preset, q.MaxRate, q.AudioBitrate)
	}
	_ = names
}

func configCmd(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	out := fs.String("out", ".framekit.json", "output path")
	_ = fs.Parse(args)
	if _, err := os.Stat(*out); err == nil {
		fatal(fmt.Errorf("%s already exists", *out))
	}
	cfg := framekit.DefaultConfig()
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Println("wrote", *out)
}
