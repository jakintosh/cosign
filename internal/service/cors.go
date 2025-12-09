package service

import (
	"strings"
)

type AllowedOrigin struct {
	URL string `json:"url"`
}

type CORSStore interface {
	CountOrigins() (int, error)
	GetOrigins() ([]AllowedOrigin, error)
	SetOrigins([]AllowedOrigin) error
}

var corsStore CORSStore

func SetCORSStore(s CORSStore) {
	corsStore = s
}

func InitCORS(
	origins []string,
) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	count, err := corsStore.CountOrigins()
	if err != nil {
		return DatabaseError{err}
	}
	if count > 0 {
		return nil
	}

	var list []AllowedOrigin
	for _, o := range origins {
		if t := strings.TrimSpace(o); t != "" {
			list = append(list, AllowedOrigin{URL: t})
		}
	}
	if len(list) == 0 {
		return nil
	}

	if err := corsStore.SetOrigins(list); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func GetAllowedOrigins() (
	[]AllowedOrigin,
	error,
) {
	if corsStore == nil {
		return nil, ErrNoCORSStore
	}

	origins, err := corsStore.GetOrigins()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return origins, nil
}

func SetAllowedOrigins(
	origins []AllowedOrigin,
) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	for _, o := range origins {
		if !isValidOrigin(o.URL) {
			return ErrCORSOriginNotFound
		}
	}

	if err := corsStore.SetOrigins(origins); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func AddAllowedOrigin(origin string) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	origin = strings.TrimSpace(origin)
	if !isValidOrigin(origin) {
		return ErrCORSOriginNotFound
	}

	origins, err := corsStore.GetOrigins()
	if err != nil {
		return DatabaseError{err}
	}

	for _, o := range origins {
		if o.URL == origin {
			return nil
		}
	}

	origins = append(origins, AllowedOrigin{URL: origin})
	if err := corsStore.SetOrigins(origins); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func DeleteAllowedOrigin(origin string) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ErrCORSOriginNotFound
	}

	origins, err := corsStore.GetOrigins()
	if err != nil {
		return DatabaseError{err}
	}

	index := -1
	for i, o := range origins {
		if o.URL == origin {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrCORSOriginNotFound
	}

	origins = append(origins[:index], origins[index+1:]...)
	if err := corsStore.SetOrigins(origins); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func IsAllowedOrigin(
	origin string,
) (
	bool,
	error,
) {
	if corsStore == nil {
		return false, ErrNoCORSStore
	}

	origins, err := corsStore.GetOrigins()
	if err != nil {
		return false, DatabaseError{err}
	}

	for _, o := range origins {
		if o.URL == origin {
			return true, nil
		}
	}
	return false, nil
}

func isValidOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}
	if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
		return true
	}
	// Allow bare hostnames/identifiers but reject other schemes
	if strings.Contains(origin, "://") {
		return false
	}
	return true
}
