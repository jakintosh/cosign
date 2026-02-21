package app

import (
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

const defaultPageSize = 10

type Options struct {
	Client   wire.Client
	PageSize int
}

type Server struct {
	client   wire.Client
	renderer *Renderer
	pageSize int
}

func New(opts Options) (*Server, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, err
	}

	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	return &Server{
		client:   opts.Client,
		renderer: renderer,
		pageSize: pageSize,
	}, nil
}

func (s *Server) Serve(addr string) error {
	return http.ListenAndServe(addr, s.BuildRouter())
}

func (s *Server) BuildRouter() http.Handler {
	mux := http.NewServeMux()

	static := http.StripPrefix("/static/", s.renderer.StaticHandler())
	mux.Handle("GET /static/", static)
	s.registerCampaignRoutes(mux)
	s.registerCampaignDetailRoutes(mux)
	s.registerLocationRoutes(mux)
	s.registerSignatureRoutes(mux)

	return withMethodOverride(mux)
}

func (s *Server) registerCampaignRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /campaigns", s.handleCampaigns)
	mux.HandleFunc("POST /campaigns", s.handleCreateCampaign)
	mux.HandleFunc("DELETE /campaigns/{campaign_id}", s.handleDeleteCampaign)
}

func (s *Server) registerCampaignDetailRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /campaigns/{campaign_id}", s.handleCampaignDetailPage)
	mux.HandleFunc("PATCH /campaigns/{campaign_id}", s.handleUpdateCampaign)
}

func (s *Server) registerLocationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /campaigns/{campaign_id}/locations", s.handleLocations)
	mux.HandleFunc("POST /campaigns/{campaign_id}/locations", s.handleCreateLocation)
	mux.HandleFunc("PATCH /campaigns/{campaign_id}/locations/{index}", s.handleUpdateLocation)
	mux.HandleFunc("DELETE /campaigns/{campaign_id}/locations/{index}", s.handleDeleteLocation)
	mux.HandleFunc("PATCH /campaigns/{campaign_id}/locations/settings", s.handleUpdateLocationsSettings)
}

func (s *Server) registerSignatureRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /campaigns/{campaign_id}/signatures", s.handleSignatures)
	mux.HandleFunc("POST /campaigns/{campaign_id}/signatures", s.handleCreateSignature)
	mux.HandleFunc("DELETE /campaigns/{campaign_id}/signatures/{signature_id}", s.handleDeleteSignature)
}

func withMethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			override := strings.ToUpper(strings.TrimSpace(r.FormValue("_method")))
			if override == http.MethodPatch || override == http.MethodDelete {
				r.Method = override
			}
		}

		next.ServeHTTP(w, r)
	})
}
