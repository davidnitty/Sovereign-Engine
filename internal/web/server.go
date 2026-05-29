package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/yourusername/the-engine/internal/app"
	"github.com/yourusername/the-engine/internal/composition"
)

type Server struct {
	app *app.App
	hub *Hub
	tpl *template.Template
}

func NewServer(a *app.App) http.Handler {
	s := &Server{
		app: a,
		hub: NewHub(),
		tpl: template.Must(template.New("home").Parse(homeTemplate)),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.Handle("/events", s.hub)
	mux.HandleFunc("/api/resources", s.resources)
	mux.HandleFunc("/api/deploy", s.deploy)
	mux.HandleFunc("/api/openapi.json", s.openapi)

	var handler http.Handler = mux
	handler = requestID(handler)
	handler = logger(handler)
	handler = rateLimit(120, time.Minute)(handler)
	return handler
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	resources, _ := s.app.Engine.List(map[string]string{})
	_ = s.tpl.Execute(w, map[string]any{
		"Resources": resources,
		"DataDir":   s.app.Config.DataDir,
	})
}

func (s *Server) resources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filter := map[string]string{}
	if provider := r.URL.Query().Get("provider"); provider != "" {
		filter["provider"] = provider
	}
	resources, err := s.app.Engine.List(filter)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, resources)
}

func (s *Server) deploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Composition string            `json:"composition"`
		Provider    string            `json:"provider"`
		Name        string            `json:"name"`
		Region      string            `json:"region"`
		Size        string            `json:"size"`
		Labels      map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	resource, err := s.app.Engine.Deploy(composition.DeployOptions{
		Composition: req.Composition,
		Provider:    req.Provider,
		Name:        req.Name,
		Region:      req.Region,
		Size:        req.Size,
		Labels:      req.Labels,
	})
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	s.hub.Broadcast(Event{Type: "resource.created", Timestamp: time.Now().UTC(), Payload: resource})
	writeJSON(w, resource)
}

func (s *Server) openapi(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"openapi": "3.0.3",
		"info": map[string]string{
			"title":   "Sovereign Engine API",
			"version": "0.1.0",
		},
		"paths": map[string]any{
			"/api/resources": map[string]any{"get": map[string]string{"summary": "List managed resources"}},
			"/api/deploy":    map[string]any{"post": map[string]string{"summary": "Deploy a composition"}},
			"/events":        map[string]any{"get": map[string]string{"summary": "Server-sent event stream"}},
		},
	})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

const homeTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sovereign Engine</title>
  <style>
    body { font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; background: #f7f7f4; color: #1f2933; }
    header { padding: 24px 32px; background: #102a43; color: #fff; }
    main { padding: 24px 32px; }
    table { width: 100%; border-collapse: collapse; background: #fff; }
    th, td { padding: 10px 12px; border-bottom: 1px solid #d9e2ec; text-align: left; font-size: 14px; }
    th { color: #52606d; }
    .meta { color: #627d98; font-size: 14px; }
  </style>
</head>
<body>
  <header>
    <h1>Sovereign Engine</h1>
    <div class="meta">Data directory: {{.DataDir}}</div>
  </header>
  <main>
    <h2>Managed Resources</h2>
    <table>
      <thead><tr><th>ID</th><th>Name</th><th>Provider</th><th>Type</th><th>Status</th><th>Region</th></tr></thead>
      <tbody>
      {{range .Resources}}
        <tr><td>{{.ID}}</td><td>{{.Name}}</td><td>{{.Provider}}</td><td>{{.Type}}</td><td>{{.Status}}</td><td>{{.Region}}</td></tr>
      {{else}}
        <tr><td colspan="6">No resources yet.</td></tr>
      {{end}}
      </tbody>
    </table>
  </main>
  <script>
    const events = new EventSource("/events");
    events.onmessage = function () { location.reload(); };
    events.addEventListener("resource.created", function () { location.reload(); });
  </script>
</body>
</html>`
