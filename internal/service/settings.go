package service

import "net/http"

func (s *Service) buildSettingsRouter(mux *http.ServeMux, mw Middleware) {
	s.keys.Router(mux, "/settings", mw.auth)
	s.cors.Router(mux, "/settings", mw.auth)
}
