package api

import (
	"net/http"
	"strings"

	"cosign/internal/service"
	"github.com/gorilla/mux"
)

func buildAdminKeysRouter(
	r *mux.Router,
) {
	r.HandleFunc("", withAuth(handlePostKey)).Methods("POST")
	r.HandleFunc("/{id}", withAuth(handleDeleteKey)).Methods("DELETE")
}

func handlePostKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	token, err := service.CreateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusCreated, token)
}

func handleDeleteKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing Key ID")
		return
	}
	if err := service.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
