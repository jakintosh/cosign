package app

import (
	"net/http"
	"strings"
)

func statusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "bad request") ||
		strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "required") {
		return http.StatusBadRequest
	}

	if strings.Contains(msg, "unauthorized") {
		return http.StatusUnauthorized
	}

	if strings.Contains(msg, "forbidden") || strings.Contains(msg, "csrf") {
		return http.StatusForbidden
	}

	if isNotFoundError(err) {
		return http.StatusNotFound
	}

	return http.StatusBadGateway
}
