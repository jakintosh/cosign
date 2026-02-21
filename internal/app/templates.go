package app

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/*.html static/*
var assets embed.FS

type Renderer struct {
	templates *template.Template
	staticFS  fs.FS
}

func NewRenderer() (*Renderer, error) {
	tmpl, err := template.ParseFS(assets, "templates/*.html")
	if err != nil {
		return nil, err
	}

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		templates: tmpl,
		staticFS:  staticFS,
	}, nil
}

func (r *Renderer) renderTemplate(
	w http.ResponseWriter,
	statusCode int,
	name string,
	data any,
) {
	var body bytes.Buffer
	if err := r.templates.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = body.WriteTo(w)
}

func (r *Renderer) StaticHandler() http.Handler {
	return http.FileServer(http.FS(r.staticFS))
}
