package framekit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CheckAssetPath enforces the configured local sandbox: allowed roots, denied
// extensions, no traversal and a size cap. All checks are local; no network.
func CheckAssetPath(p string, sec Security) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("empty asset path")
	}
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if part == ".." {
			return fmt.Errorf("path traversal is not allowed: %s", p)
		}
	}
	ext := strings.ToLower(filepath.Ext(p))
	for _, deny := range sec.DenyExtensions {
		if ext == strings.ToLower(deny) {
			return fmt.Errorf("extension %s is denied", ext)
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	allowed := false
	for _, root := range sec.AllowedAssetRoots {
		r, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(r, abs)
		if err == nil && !strings.HasPrefix(rel, "..") {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("asset %s is outside allowed roots", p)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("asset not found: %s", p)
	}
	if info.Size() > int64(sec.MaxAssetSizeMB*1024*1024) {
		return fmt.Errorf("asset %s exceeds %.1f MB cap", p, sec.MaxAssetSizeMB)
	}
	return nil
}

// SanitizeText removes control characters and collapses blank lines so agent
// generated copy cannot inject terminal escapes into the renderer.
func SanitizeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsControl(r) {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	lines := strings.Split(b.String(), "\n")
	out := lines[:0]
	blank := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			if blank {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// SanitizeSpec applies text hygiene to every user-facing string.
func SanitizeSpec(spec *Spec) {
	spec.Title = SanitizeText(spec.Title)
	for i := range spec.Scenes {
		sc := &spec.Scenes[i]
		sc.Title = SanitizeText(sc.Title)
		sc.Subtitle = SanitizeText(sc.Subtitle)
		sc.Name = SanitizeText(sc.Name)
		sc.Command = SanitizeText(sc.Command)
		for j := range sc.Lines {
			sc.Lines[j].Text = SanitizeText(sc.Lines[j].Text)
		}
		for j := range sc.Items {
			sc.Items[j].Name = SanitizeText(sc.Items[j].Name)
			sc.Items[j].Note = SanitizeText(sc.Items[j].Note)
		}
	}
	for i := range spec.Voices {
		spec.Voices[i].Text = SanitizeText(spec.Voices[i].Text)
	}
}
