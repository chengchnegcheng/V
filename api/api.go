package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"v/errors"
	"v/logger"
	"v/middleware"
)

// Handler represents an API handler
type Handler struct {
	log      *logger.Logger
	router   *mux.Router
	handlers map[string]http.HandlerFunc
}

// New creates a new API handler
func New(log *logger.Logger) *Handler {
	return &Handler{
		log:      log,
		router:   mux.NewRouter(),
		handlers: make(map[string]http.HandlerFunc),
	}
}

// Register registers a new handler
func (h *Handler) Register(path string, handler http.HandlerFunc) {
	h.handlers[path] = handler
}

// Setup sets up the API routes
func (h *Handler) Setup() {
	// Add middleware
	h.router.Use(middleware.Logging(h.log))
	h.router.Use(middleware.Recovery(h.log))
	h.router.Use(middleware.CORS())
	h.router.Use(middleware.RateLimit())

	// Register handlers
	for path, handler := range h.handlers {
		h.router.HandleFunc(path, handler)
	}

	// Add not found handler
	h.router.NotFoundHandler = http.HandlerFunc(h.handleNotFound)
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// handleNotFound handles 404 errors
func (h *Handler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	h.handleError(w, errors.New(errors.ErrNotFound, "Not found", nil))
}

// handleError handles API errors
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var apiErr *errors.Error
	if e, ok := err.(*errors.Error); ok {
		apiErr = e
	} else {
		apiErr = errors.New(errors.ErrInternalServer, err.Error(), err)
	}

	// Log error
	h.log.Error("API error", logger.Fields{
		"code":    apiErr.Code,
		"message": apiErr.Message,
		"error":   apiErr.Err,
	})

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	json.NewEncoder(w).Encode(apiErr)
}

// handleResponse handles API responses
func (h *Handler) handleResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getPathParam gets a path parameter
func (h *Handler) getPathParam(r *http.Request, name string) string {
	return mux.Vars(r)[name]
}

// getQueryParam gets a query parameter
func (h *Handler) getQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// getAuthToken gets the authentication token
func (h *Handler) getAuthToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// getContentType gets the content type
func (h *Handler) getContentType(r *http.Request) string {
	return r.Header.Get("Content-Type")
}

// getUserAgent gets the user agent
func (h *Handler) getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// getIP gets the client IP
func (h *Handler) getIP(r *http.Request) string {
	// Try X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Try X-Forwarded-For header
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	// Use remote address
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
