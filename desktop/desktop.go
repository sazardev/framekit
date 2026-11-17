// Package desktop provides a local web studio for framekit.
package desktop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/sazardev/framekit"
)

type job struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Path    string `json:"-"`
}

var (
	jobsMu sync.Mutex
	jobs   = map[string]*job{}
)

// Run starts the local studio. It binds 127.0.0.1 only.
func Run(port int, noOpen bool) error {
	cfg, err := framekit.LoadConfig("")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, desktopHTML)
	})
	mux.HandleFunc("/api/lint", func(w http.ResponseWriter, r *http.Request) {
		spec, q, err := decodeRequest(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		a, err := framekit.Analyze(spec, cfg, q)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		findings := append(framekit.Lint(spec, cfg, q, a), framekit.SecurityAudit(spec, cfg)...)
		writeJSON(w, 200, map[string]any{"analysis": a, "findings": findings})
	})
	mux.HandleFunc("/api/render", func(w http.ResponseWriter, r *http.Request) {
		spec, q, err := decodeRequest(r)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		j := &job{Status: "queued", Message: "starting…"}
		jobsMu.Lock()
		jobs[id] = j
		jobsMu.Unlock()
		dir, err := os.MkdirTemp("", "framekit-desktop-*")
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		out := filepath.Join(dir, "video.mp4")
		go func() {
			updateJob(id, "rendering", "encoding 1080p60…", out)
			if err := framekit.RenderWithConfig(spec, out, cfg, q); err != nil {
				updateJob(id, "error", err.Error(), out)
				return
			}
			updateJob(id, "done", "ready", out)
		}()
		writeJSON(w, 200, map[string]any{"id": id})
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		jobsMu.Lock()
		j, ok := jobs[id]
		jobsMu.Unlock()
		if !ok {
			writeJSON(w, 404, map[string]any{"error": "unknown job"})
			return
		}
		writeJSON(w, 200, j)
	})
	mux.HandleFunc("/api/video", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		jobsMu.Lock()
		j, ok := jobs[id]
		jobsMu.Unlock()
		if !ok || j.Path == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeFile(w, r, j.Path)
	})

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("framekit desktop → %s\n", url)
	if !noOpen {
		openBrowser(url)
	}
	return http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}

func updateJob(id, status, message, path string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if j, ok := jobs[id]; ok {
		j.Status, j.Message, j.Path = status, message, path
	}
}

func decodeRequest(r *http.Request) (framekit.Spec, string, error) {
	var body struct {
		Spec    string `json:"spec"`
		Quality string `json:"quality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return framekit.Spec{}, "", err
	}
	var spec framekit.Spec
	if err := json.Unmarshal([]byte(body.Spec), &spec); err != nil {
		return framekit.Spec{}, "", err
	}
	framekit.SanitizeSpec(&spec)
	return spec, body.Quality, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start()
}
