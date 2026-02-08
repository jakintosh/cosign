package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~jakintosh/command-go/pkg/cors"
	"git.sr.ht/~jakintosh/command-go/pkg/keys"
	"git.sr.ht/~jakintosh/command-go/pkg/wire"
	"golang.org/x/time/rate"
)

var (
	ErrCampaignNotFound     = errors.New("campaign not found")
	ErrSignatureNotFound    = errors.New("signature not found")
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrDuplicateEmail       = errors.New("email already signed")
	ErrLocationNotInOptions = errors.New("location must be from preset options")
	ErrEmptyName            = errors.New("name cannot be empty")
	ErrEmptyEmail           = errors.New("email cannot be empty")
	ErrEmptyLocation        = errors.New("location cannot be empty")
	ErrEmptyCampaignName    = errors.New("campaign name cannot be empty")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }

type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AllowCustomText bool   `json:"allow_custom_text"`
	CreatedAt       int64  `json:"created_at"`
}

type LocationOption struct {
	ID           int64  `json:"id,omitempty"`
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

type Campaigns struct {
	Campaigns []*Campaign `json:"campaigns"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}

type Signature struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Location  string `json:"location"`
	CreatedAt int64  `json:"created_at"`
}

type Signatures struct {
	Signatures []*Signature `json:"signatures"`
	Total      int          `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type Store interface {
	InsertCampaign(id, name string, allowCustomText bool, createdAt int64) error
	GetCampaign(id string) (*Campaign, error)
	ListCampaigns(limit, offset int) ([]*Campaign, error)
	CountCampaigns() (int, error)
	UpdateCampaign(id, name string, allowCustomText bool) error
	DeleteCampaign(id string) error
	GetCampaignLocations(campaignID string) ([]*LocationOption, error)
	ReplaceCampaignLocations(campaignID string, options []LocationOption) error

	InsertSignature(campaignID, name, email, location string, createdAt int64) (int64, error)
	GetSignature(campaignID string, id int64) (*Signature, error)
	ListSignatures(campaignID string, limit, offset int) ([]*Signature, error)
	CountSignatures(campaignID string) (int, error)
	DeleteSignature(campaignID string, id int64) error
	SignatureEmailExists(campaignID, email string) (bool, error)
}

type Options struct {
	Store       Store
	KeysOptions *keys.Options
	CORSOptions *cors.Options
	Clock       func() time.Time
	HealthCheck func() error
}

type Service struct {
	store       Store
	keys        *keys.Service
	cors        *cors.Service
	clock       func() time.Time
	healthCheck func() error

	rateLimiters   map[string]*rate.Limiter
	rateLimitersMu sync.Mutex
}

func New(opts Options) (*Service, error) {
	if opts.Store == nil {
		return nil, errors.New("service: store required")
	}
	if opts.KeysOptions == nil {
		return nil, errors.New("service: keys options required")
	}
	if opts.CORSOptions == nil {
		return nil, errors.New("service: cors options required")
	}

	keysSvc, err := keys.New(*opts.KeysOptions)
	if err != nil {
		return nil, err
	}

	corsSvc, err := cors.New(*opts.CORSOptions)
	if err != nil {
		return nil, err
	}

	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}

	healthCheck := opts.HealthCheck
	if healthCheck == nil {
		healthCheck = func() error { return nil }
	}

	return &Service{
		store:        opts.Store,
		keys:         keysSvc,
		cors:         corsSvc,
		clock:        clock,
		healthCheck:  healthCheck,
		rateLimiters: make(map[string]*rate.Limiter),
	}, nil
}

func (s *Service) Serve(
	addr string,
	apiPrefix string,
) error {
	apiHandler := http.StripPrefix(apiPrefix, s.BuildRouter())
	rootMux := http.NewServeMux()
	rootMux.Handle(apiPrefix+"/", apiHandler)
	rootMux.Handle(apiPrefix, apiHandler)
	return http.ListenAndServe(addr, rootMux)
}

func (s *Service) withRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := s.rateLimiterFor(ip)
		if !limiter.Allow() {
			wire.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next(w, r)
	}
}

func (s *Service) rateLimiterFor(ip string) *rate.Limiter {
	s.rateLimitersMu.Lock()
	defer s.rateLimitersMu.Unlock()

	limiter, ok := s.rateLimiters[ip]
	if ok {
		return limiter
	}

	limiter = rate.NewLimiter(10, 20)
	s.rateLimiters[ip] = limiter
	return limiter
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.healthCheck(); err != nil {
		wire.WriteError(w, http.StatusServiceUnavailable, "database unhealthy")
		return
	}

	wire.WriteData(w, http.StatusOK, HealthResponse{Status: "healthy"})
}

func campaignIDFromPath(r *http.Request) string {
	return strings.TrimSpace(r.PathValue("campaign_id"))
}

func signatureIDFromPath(r *http.Request) (int64, error) {
	v := strings.TrimSpace(r.PathValue("signature_id"))
	if v == "" {
		return 0, fmt.Errorf("missing signature id")
	}
	return strconv.ParseInt(v, 10, 64)
}

func randomID(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
