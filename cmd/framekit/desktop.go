package main

import (
	"encoding/json"
	"flag"
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

func desktopCmd(args []string) {
	fs := flag.NewFlagSet("desktop", flag.ExitOnError)
	port := fs.Int("port", 8239, "local port")
	noOpen := fs.Bool("no-open", false, "do not open the browser")
	_ = fs.Parse(args)

	cfg := loadCfg("")
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

	url := fmt.Sprintf("http://127.0.0.1:%d", *port)
	fmt.Printf("framekit desktop → %s\n", url)
	if !*noOpen {
		openBrowser(url)
	}
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *port), mux); err != nil {
		fatal(err)
	}
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

const desktopHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Framekit Studio</title>
<style>
:root{--bg:#282828;--deep:#202020;--soft:#323232;--fg:#ebdbb2;--dim:#bdae93;--faint:#665c54;--green:#b8bb26;--aqua:#8ec07c}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:14px ui-monospace,SFMono-Regular,Menlo,monospace;padding:24px}
header{display:flex;align-items:center;gap:12px;border-bottom:1px solid var(--faint);padding-bottom:14px;margin-bottom:22px}
.sq{width:14px;height:14px;background:var(--green)}
h1{font-size:18px;margin:0;letter-spacing:1px}
small{color:var(--faint)}
.grid{display:grid;grid-template-columns:1fr 360px;gap:22px}
textarea{width:100%;height:560px;background:var(--deep);color:var(--fg);border:1px solid var(--faint);padding:14px;font:12px/1.5 ui-monospace,monospace;resize:vertical}
textarea:focus{outline:none;border-color:var(--green)}
.panel{border:1px solid var(--faint);padding:14px;margin-bottom:14px}
label{display:block;color:var(--dim);margin-bottom:6px;font-size:12px}
select{width:100%;background:var(--deep);color:var(--fg);border:1px solid var(--faint);padding:8px}
button{background:var(--deep);color:var(--fg);border:1px solid var(--faint);padding:10px 14px;cursor:pointer;width:100%;margin-top:10px}
button:hover{border-color:var(--green);color:var(--green)}
button.primary{background:var(--green);color:var(--deep);border-color:var(--green);font-weight:bold}
button.primary:hover{opacity:.85;color:var(--deep)}
pre{white-space:pre-wrap;word-break:break-word;color:var(--dim);font-size:11px;max-height:200px;overflow:auto}
.ok{color:var(--aqua)} .err{color:#f29b7b} video{width:100%;margin-top:12px;border:1px solid var(--faint)}
</style></head><body>
<header><div class="sq"></div><h1>FRAMEKIT</h1><small>studio · 1080p60 · local render</small></header>
<div class="grid">
 <div><label>SPEC (JSON)</label><textarea id="spec"></textarea></div>
 <div>
  <div class="panel"><label>QUALITY</label><select id="quality">
   <option>draft</option><option>standard</option><option>social</option><option selected>high</option><option>master</option>
  </select><button onclick="doLint()">Lint & analyze</button><button class="primary" onclick="doRender()">Render video</button></div>
  <div class="panel"><label>STATUS</label><div id="status">idle</div><pre id="report"></pre></div>
  <div class="panel"><label>PREVIEW</label><div id="preview"></div></div>
 </div>
</div>
<script>
const TEMPLATE={
 "title":"SazarDev — Freelance Developer Intro","format":"landscape","fps":60,
 "effects":{"typewriter":true,"cursor":true,"chrome":true,"progress":true,"fade":0.38,"slide":10,"easing":"easeOutQuint","stagger":0.4},
 "music":{"enabled":true,"volume":0.25,"progression":"ambient","tempo":80},
 "scenes":[
  {"kind":"title","name":"01 / IDENTITY","duration":3.5,"title":"Omar Flores","subtitle":"Full-Stack Freelance Developer"},
  {"kind":"terminal","name":"02 / STACK","duration":5,"command":"cat stack.txt","items":[{"name":"Flutter / Dart","note":"mobile"},{"name":"React / TypeScript","note":"web"},{"name":"Go / Rust","note":"systems"}]},
  {"kind":"stats","name":"03 / IMPACT","duration":4.5,"stats":[{"value":"29","label":"APPS"},{"value":"26","label":"PROJECTS"},{"value":"300+","label":"STARS"}]},
  {"kind":"credits","name":"04 / CONTACT","duration":4,"title":"Let's build","subtitle":"@salazarprograms · Fiverr"}],
 "output":"studio.mp4"};
document.getElementById('spec').value=JSON.stringify(TEMPLATE,null,2);
const $=id=>document.getElementById(id);
const payload=()=>JSON.stringify({spec:$('spec').value,quality:$('quality').value});
async function doLint(){$('status').textContent='checking…';
 const r=await fetch('/api/lint',{method:'POST',body:payload()});const j=await r.json();
 if(j.error){$('status').innerHTML='<span class=err>'+j.error+'</span>';return;}
 const e=(j.findings||[]).filter(f=>f.severity==='error').length;
 $('status').innerHTML=e?'<span class=err>'+e+' error(s)</span>':'<span class=ok>clean</span>';
 $('report').textContent=JSON.stringify(j.findings,null,2)+'\n\n'+JSON.stringify(j.analysis,null,2).slice(0,900);}
async function doRender(){$('status').textContent='rendering…';$('preview').innerHTML='';
 const r=await fetch('/api/render',{method:'POST',body:payload()});const j=await r.json();
 if(j.error){$('status').innerHTML='<span class=err>'+j.error+'</span>';return;}
 const id=j.id;const tick=async()=>{const s=await (await fetch('/api/status?id='+id)).json();
  $('status').textContent=s.status+' · '+s.message;
  if(s.status==='done'){$('status').innerHTML='<span class=ok>ready</span>';
    $('preview').innerHTML='<video controls src="/api/video?id='+id+'"></video><a href="/api/video?id='+id+'" download>download</a>';return;}
  if(s.status==='error'){$('status').innerHTML='<span class=err>'+s.message+'</span>';return;}
  setTimeout(tick,700);};tick();}
</script></body></html>`
