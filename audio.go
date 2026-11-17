package framekit

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const sampleRate = 44100

// progressionChords maps a named mood to its chord frequencies. All original,
// all synthesized locally. "ambient" is the portfolio intro sound.
func progressionChords(name string) [][]float64 {
	switch name {
	case "uplift":
		return [][]float64{
			{261.63, 329.63, 392.00, 493.88}, // Cmaj7
			{349.23, 440.00, 523.25, 659.25}, // Fmaj7
			{392.00, 493.88, 587.33, 698.46}, // Gmaj7
			{329.63, 392.00, 493.88, 587.33}, // Emin7
		}
	case "noir":
		return [][]float64{
			{146.83, 174.61, 220.00, 261.63}, // Dm7
			{130.81, 155.56, 196.00, 233.08}, // Cm7
			{110.00, 130.81, 164.81, 196.00}, // Am7
			{123.47, 146.83, 185.00, 220.00}, // Bdim-ish
		}
	case "minimal":
		return [][]float64{
			{220.00, 329.63, 440.00},
			{196.00, 293.66, 392.00},
			{174.61, 261.63, 349.23},
			{196.00, 293.66, 392.00},
		}
	default: // ambient
		return [][]float64{
			{220.00, 261.63, 329.63, 392.00}, // Am7
			{174.61, 220.00, 261.63, 329.63}, // Fmaj7
			{261.63, 329.63, 392.00, 493.88}, // Cmaj7
			{196.00, 246.94, 293.66, 392.00}, // G6
		}
	}
}

func progressionNotes(name string) []float64 {
	switch name {
	case "uplift":
		return []float64{523.25, 659.25, 783.99, 987.77, 783.99, 659.25, 587.33, 493.88}
	case "noir":
		return []float64{349.23, 415.30, 523.25, 440.00, 392.00, 349.23, 311.13, 261.63}
	case "minimal":
		return []float64{440.00, 659.25, 587.33, 440.00, 392.00, 587.33, 523.25, 392.00}
	default:
		return []float64{440.00, 523.25, 659.25, 783.99, 659.25, 523.25, 493.88, 392.00}
	}
}

// SynthMusic generates a deterministic ambient WAV locally: a four-chord pad,
// a decaying arpeggio and soft noise ticks. No samples, no network, no licenses.
func SynthMusic(dur float64, cfg Music) ([]byte, error) {
	if dur <= 0 {
		return nil, fmt.Errorf("invalid duration %v", dur)
	}
	n := int(dur * sampleRate)
	mix := make([]float64, n)
	chords := progressionChords(cfg.Progression)
	segDur := dur / float64(len(chords))
	segN := int(segDur * sampleRate)
	for ci, chord := range chords {
		start := ci * segN
		for i := 0; i < segN && start+i < n; i++ {
			sec := float64(i) / sampleRate
			env := math.Min(sec/1.3, 1) * math.Min((segDur-sec)/1.3, 1)
			if env < 0 {
				env = 0
			}
			detune := 0.004 * math.Sin(2*math.Pi*0.11*sec)
			var v float64
			for _, f := range chord {
				v += math.Sin(2 * math.Pi * (f + detune) * sec)
			}
			mix[start+i] += v / float64(len(chord)) * env * 0.20
		}
	}

	notes := progressionNotes(cfg.Progression)
	tempo := cfg.Tempo
	if tempo <= 0 {
		tempo = 80
	}
	step := 60 / tempo / 2 // eighth notes
	for k := 0; k < int(dur/step); k++ {
		f := notes[k%len(notes)]
		st := int(float64(k) * step * sampleRate)
		ln := int(0.5 * sampleRate)
		if st+ln > n {
			ln = n - st
		}
		if ln <= 0 {
			break
		}
		for i := 0; i < ln; i++ {
			t := float64(i) / sampleRate
			env := math.Exp(-t*5.5) * (1 - math.Exp(-t*90))
			tone := math.Sin(2*math.Pi*f*t) + 0.25*math.Sin(2*math.Pi*2*f*t)
			mix[st+i] += tone * env * 0.11
		}
	}

	rng := rand.New(rand.NewSource(7))
	for k := 0; k < int(dur/0.75); k++ {
		st := int(float64(k) * 0.75 * sampleRate)
		ln := int(35 * sampleRate / 1000)
		if st+ln > n {
			ln = n - st
		}
		if ln <= 0 {
			break
		}
		for i := 0; i < ln; i++ {
			env := math.Exp(-float64(i) / (0.007 * sampleRate))
			mix[st+i] += rng.NormFloat64() * env * 0.018
		}
	}

	fi, fo := int(1.6*sampleRate), int(2.6*sampleRate)
	for i := 0; i < fi && i < n; i++ {
		mix[i] *= float64(i) / float64(fi)
	}
	for i := 0; i < fo && i < n; i++ {
		mix[n-1-i] *= float64(i) / float64(fo)
	}
	peak := 0.0
	for _, v := range mix {
		if math.Abs(v) > peak {
			peak = math.Abs(v)
		}
	}
	// Volume is the final mix level in [0,1]; the synth normalizes to a
	// musical peak first, then scales so 0.25 is a pleasant background bed.
	target := cfg.Volume * 2.4
	if target > 1 {
		target = 1
	}
	gain := 0.0
	if peak > 0 {
		gain = target / peak
	}
	pcm := make([]int16, n)
	for i, v := range mix {
		pcm[i] = int16(math.Max(-1, math.Min(1, v*gain)) * 32767)
	}
	return wav(pcm), nil
}

func wav(samples []int16) []byte {
	var buf bytes.Buffer
	dataLen := len(samples) * 2
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVEfmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(dataLen))
	_ = binary.Write(&buf, binary.LittleEndian, samples)
	return buf.Bytes()
}

// resolveVoice returns a local path for a voice, downloading it from
// Cloudflare Workers AI when provider is "cloudflare". Credentials come from
// CF_ACCOUNT_ID and CF_API_TOKEN; nothing is sent unless a voice is declared.
func resolveVoice(v Voice) (string, error) {
	if v.Provider == "local" || v.File != "" {
		if _, err := os.Stat(v.File); err != nil {
			return "", fmt.Errorf("voice file: %w", err)
		}
		return v.File, nil
	}
	if strings.TrimSpace(v.Text) == "" {
		return "", nil
	}
	account := os.Getenv("CF_ACCOUNT_ID")
	token := os.Getenv("CF_API_TOKEN")
	if account == "" || token == "" {
		return "", fmt.Errorf("cloudflare voice requires CF_ACCOUNT_ID and CF_API_TOKEN")
	}
	model := v.Model
	if model == "" {
		model = "@cf/facebook/mms-tts-eng"
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", account, model)
	body, _ := json.Marshal(map[string]string{"text": v.Text})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("cloudflare %s: %s", resp.Status, string(data))
	}
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
		var env struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if json.Unmarshal(data, &env) == nil && len(env.Errors) > 0 {
			return "", fmt.Errorf("cloudflare: %s", env.Errors[0].Message)
		}
	}
	f, err := os.CreateTemp("", "framekit-voice-*")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func mux(ffmpeg, silent string, audios []string, spec Spec, out string) error {
	args := []string{"-y", "-loglevel", "error", "-i", silent}
	for _, a := range audios {
		args = append(args, "-i", a)
	}
	var filters []string
	for i, v := range spec.Voices {
		idx := i + 2 // 0 = video, 1 = music (if present)
		if !spec.Music.Enabled {
			idx = i + 1
		}
		delay := int(v.At * 1000)
		filters = append(filters, fmt.Sprintf("[%d:a]adelay=%d|%d,volume=%.2f[v%d]", idx, delay, delay, v.Volume, i))
	}
	mapAudio := "1:a"
	if len(filters) > 0 {
		inputs := []string{"1:a"}
		for i := range filters {
			inputs = append(inputs, fmt.Sprintf("[v%d]", i))
		}
		filters = append(filters, fmt.Sprintf("%samix=inputs=%d:duration=first[aout]", strings.Join(inputs, ""), len(inputs)))
		args = append(args, "-filter_complex", strings.Join(filters, ";"))
		mapAudio = "[aout]"
	}
	args = append(args,
		"-map", "0:v", "-map", mapAudio,
		"-c:v", "copy", "-c:a", "aac", "-b:a", "128k",
		"-shortest", "-movflags", "+faststart", out)
	cmd := exec.Command(ffmpeg, args...)
	msg, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg mux: %w: %s", err, strings.TrimSpace(string(msg)))
	}
	return nil
}
