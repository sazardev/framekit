package framekit

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// Generate produces deterministic variants of a spec. It is useful for A/B
// testing creatives and for agents that need several options from one brief.
func Generate(base Spec, count int, seed int64) ([]Spec, error) {
	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20
	}
	rng := rand.New(rand.NewSource(seed))
	easings := []string{"easeOutCubic", "easeOutQuint", "easeInOut", "spring"}
	progressions := []string{"ambient", "uplift", "noir", "minimal"}
	themes := []string{"gruvbox", "catppuccin", "tokyo-night", "nord", "dracula"}

	out := make([]Spec, 0, count)
	for i := 0; i < count; i++ {
		b, err := json.Marshal(base)
		if err != nil {
			return nil, err
		}
		var s Spec
		if err := json.Unmarshal(b, &s); err != nil {
			return nil, err
		}
		s.Title = fmt.Sprintf("%s · v%d", s.Title, i+1)
		s.Effects.Easing = easings[rng.Intn(len(easings))]
		if s.Music.Enabled {
			s.Music.Progression = progressions[rng.Intn(len(progressions))]
			s.Music.Tempo = 70 + float64(rng.Intn(30))
		}
		s.Theme = &Theme{Name: themes[rng.Intn(len(themes))]}
		s.Effects.Transition = []string{"bgfade", "wipe"}[rng.Intn(2)]
		s.Effects.Scanline = rng.Intn(2) == 0
		s.Effects.Tracking = 1 + rng.Float64()*2
		for j := range s.Scenes {
			jitter := 0.8 + rng.Float64()*0.45
			s.Scenes[j].Duration *= jitter
		}
		SanitizeSpec(&s)
		out = append(out, s)
	}
	return out, nil
}
