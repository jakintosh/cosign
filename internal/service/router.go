package service

import "net/http"

type Middleware struct {
	auth      func(http.HandlerFunc) http.HandlerFunc
	cors      func(http.HandlerFunc) http.HandlerFunc
	rateLimit func(http.HandlerFunc) http.HandlerFunc
}

func (s *Service) BuildRouter() http.Handler {
	mux := http.NewServeMux()
	mw := Middleware{
		auth:      s.keys.WithAuth,
		cors:      s.cors.WithCORS,
		rateLimit: s.withRateLimit,
	}

	s.buildHealthRouter(mux)
	s.buildPublicRouter(mux, mw)
	s.buildAdminRouter(mux, mw)
	s.buildSettingsRouter(mux, mw)

	return mux
}

func (s *Service) buildHealthRouter(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth)
}

func (s *Service) buildPublicRouter(mux *http.ServeMux, mw Middleware) {
	campaignsMux := http.NewServeMux()
	s.buildPublicCampaignRouter(campaignsMux, mw)
	s.buildPublicSignonRouter(campaignsMux, mw)

	mountSubrouter(mux, "/campaigns", campaignsMux)
}

func (s *Service) buildAdminRouter(mux *http.ServeMux, mw Middleware) {
	adminMux := http.NewServeMux()
	s.buildAdminCampaignsRouter(adminMux, mw)
	securedAdmin := http.HandlerFunc(mw.auth(adminMux.ServeHTTP))

	mountSubrouter(mux, "/admin", securedAdmin)
}

func (s *Service) buildAdminCampaignsRouter(mux *http.ServeMux, mw Middleware) {
	campaignsMux := http.NewServeMux()
	s.buildAdminCampaignRouter(campaignsMux, mw)
	s.buildAdminSignonRouter(campaignsMux, mw)

	mountSubrouter(mux, "/campaigns", campaignsMux)
}

func mountSubrouter(parent *http.ServeMux, prefix string, child http.Handler) {
	stripped := http.StripPrefix(prefix, child)
	parent.Handle(prefix+"/", stripped)
	parent.Handle(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(r.Context())
		u := *r.URL
		u.Path = prefix + "/"
		if r.URL.RawPath != "" {
			u.RawPath = prefix + "/"
		}
		r2.URL = &u
		stripped.ServeHTTP(w, r2)
	}))
}
