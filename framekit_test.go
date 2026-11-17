package framekit

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestNormalize(t *testing.T) {
	s := Spec{
		Format: "portrait",
		Scenes: []Scene{
			{Kind: "terminal", Duration: 2},
			{Kind: "stats", Duration: 3},
		},
	}
	if err := s.Normalize(); err != nil {
		t.Fatal(err)
	}
	if s.FPS != 30 {
		t.Fatalf("fps = %d", s.FPS)
	}
	if s.DurationSeconds() != 5 {
		t.Fatalf("duration = %v", s.DurationSeconds())
	}
	if s.Size.W != 1080 || s.Size.H != 1920 {
		t.Fatalf("size = %+v", s.Size)
	}
	if s.Theme.Name != "gruvbox" {
		t.Fatalf("theme = %q", s.Theme.Name)
	}
}

func TestPresets(t *testing.T) {
	for _, name := range []string{"landscape", "square", "portrait"} {
		if _, ok := Presets[name]; !ok {
			t.Fatalf("missing preset %s", name)
		}
	}
	s := Spec{Format: "nope", Scenes: []Scene{{Duration: 1}}}
	if err := s.Normalize(); err == nil {
		t.Fatal("expected error for bad format")
	}
}

func TestThemeMergeAndColors(t *testing.T) {
	base := Gruvbox()
	custom := &Theme{Green: "#00ff00", Brand: "Custom"}
	merged := base.Merge(custom)
	if merged.Green != "#00ff00" || merged.Brand != "Custom" {
		t.Fatalf("merge failed: %+v", merged)
	}
	if merged.BG != base.BG {
		t.Fatal("merge overwrote BG")
	}
	c, err := ParseHex("#b8bb26")
	if err != nil || c.R != 0xb8 || c.G != 0xbb || c.B != 0x26 || c.A != 255 {
		t.Fatalf("ParseHex = %+v %v", c, err)
	}
}

func TestSynthMusic(t *testing.T) {
	cfg := Music{Enabled: true, Volume: 0.25}
	a, err := SynthMusic(4, cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SynthMusic(4, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("synth is not deterministic")
	}
	if string(a[:4]) != "RIFF" || string(a[8:12]) != "WAVE" {
		t.Fatal("bad WAV header")
	}
	want := 36 + 4*sampleRate*2
	got := int(binary.LittleEndian.Uint32(a[4:8]))
	if got != want {
		t.Fatalf("RIFF size = %d, want %d", got, want)
	}
}

func TestTypedAndReveal(t *testing.T) {
	if got := typed("hello", 0.5, 1, 10); got != "" {
		t.Fatalf("typed early = %q", got)
	}
	if got := typed("hello", 2, 1, 10); got != "hello" {
		t.Fatalf("typed done = %q", got)
	}
	if got := reveal(1, 0, 0.5); got != 1 {
		t.Fatalf("reveal done = %v", got)
	}
	if got := reveal(0.1, 0, 0.5); got <= 0 || got >= 1 {
		t.Fatalf("reveal mid = %v", got)
	}
}

func TestEmbeddedFonts(t *testing.T) {
	for _, w := range []string{"regular", "bold"} {
		b, err := FontBytes(w)
		if err != nil || len(b) < 1000 {
			t.Fatalf("font %s: %v (%d bytes)", w, err, len(b))
		}
	}
}
