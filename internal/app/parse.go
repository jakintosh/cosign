package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func parsePageQuery(r *http.Request, key string) int {
	raw := strings.TrimSpace(r.FormValue(key))
	if raw == "" {
		return 1
	}

	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}

	return page
}

func parseLocationsMode(r *http.Request, modeKey, indexKey string) (string, int) {
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(modeKey)))
	if mode != "new" && mode != "edit" {
		return "", -1
	}

	if mode == "new" {
		return mode, -1
	}

	indexRaw := strings.TrimSpace(r.URL.Query().Get(indexKey))
	index, err := strconv.Atoi(indexRaw)
	if err != nil || index < 0 {
		return "", -1
	}

	return mode, index
}

func parsePathIndex(r *http.Request, pathKey string) (int, error) {
	indexRaw := strings.TrimSpace(r.PathValue(pathKey))
	index, err := strconv.Atoi(indexRaw)
	if err != nil || index < 0 {
		return 0, fmt.Errorf("invalid index")
	}

	return index, nil
}

func campaignIDFromPath(r *http.Request) string {
	return strings.TrimSpace(r.PathValue("campaign_id"))
}
