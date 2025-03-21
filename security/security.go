package security

import (
	"fmt"
	"net"
	"sync"
	"time"

	"v/logger"
)

// IPFilter handles IP filtering
type IPFilter struct {
	whitelist map[string]bool
	blacklist map[string]bool
	mu        sync.RWMutex
}

// RateLimiter handles rate limiting
type RateLimiter struct {
	limits    map[string]*rateLimit
	mu        sync.RWMutex
	cleanupCh chan struct{}
}

type rateLimit struct {
	count     int
	lastReset time.Time
}

// Security handles security features
type Security struct {
	log         *logger.Logger
	ipFilter    *IPFilter
	rateLimiter *RateLimiter
}

// New creates a new security manager
func New(log *logger.Logger) *Security {
	s := &Security{
		log: log,
		ipFilter: &IPFilter{
			whitelist: make(map[string]bool),
			blacklist: make(map[string]bool),
		},
		rateLimiter: &RateLimiter{
			limits:    make(map[string]*rateLimit),
			cleanupCh: make(chan struct{}),
		},
	}

	go s.cleanupLoop()
	return s
}

// Close closes the security manager
func (s *Security) Close() error {
	close(s.rateLimiter.cleanupCh)
	return nil
}

// AddToWhitelist adds an IP to the whitelist
func (s *Security) AddToWhitelist(ip string) error {
	s.ipFilter.mu.Lock()
	defer s.ipFilter.mu.Unlock()

	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	s.ipFilter.whitelist[ip] = true
	s.log.Info("Added IP to whitelist", logger.Fields{"ip": ip})
	return nil
}

// RemoveFromWhitelist removes an IP from the whitelist
func (s *Security) RemoveFromWhitelist(ip string) error {
	s.ipFilter.mu.Lock()
	defer s.ipFilter.mu.Unlock()

	if _, exists := s.ipFilter.whitelist[ip]; !exists {
		return fmt.Errorf("IP %s not in whitelist", ip)
	}

	delete(s.ipFilter.whitelist, ip)
	s.log.Info("Removed IP from whitelist", logger.Fields{"ip": ip})
	return nil
}

// AddToBlacklist adds an IP to the blacklist
func (s *Security) AddToBlacklist(ip string) error {
	s.ipFilter.mu.Lock()
	defer s.ipFilter.mu.Unlock()

	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	s.ipFilter.blacklist[ip] = true
	s.log.Info("Added IP to blacklist", logger.Fields{"ip": ip})
	return nil
}

// RemoveFromBlacklist removes an IP from the blacklist
func (s *Security) RemoveFromBlacklist(ip string) error {
	s.ipFilter.mu.Lock()
	defer s.ipFilter.mu.Unlock()

	if _, exists := s.ipFilter.blacklist[ip]; !exists {
		return fmt.Errorf("IP %s not in blacklist", ip)
	}

	delete(s.ipFilter.blacklist, ip)
	s.log.Info("Removed IP from blacklist", logger.Fields{"ip": ip})
	return nil
}

// IsIPAllowed checks if an IP is allowed
func (s *Security) IsIPAllowed(ip string) bool {
	s.ipFilter.mu.RLock()
	defer s.ipFilter.mu.RUnlock()

	if s.ipFilter.whitelist[ip] {
		return true
	}

	if s.ipFilter.blacklist[ip] {
		return false
	}

	return true
}

// SetRateLimit sets a rate limit for a key
func (s *Security) SetRateLimit(key string, limit int, window time.Duration) error {
	s.rateLimiter.mu.Lock()
	defer s.rateLimiter.mu.Unlock()

	s.rateLimiter.limits[key] = &rateLimit{
		count:     0,
		lastReset: time.Now(),
	}

	s.log.Info("Set rate limit", logger.Fields{
		"key":    key,
		"limit":  limit,
		"window": window,
	})

	return nil
}

// CheckRateLimit checks if a request is within rate limits
func (s *Security) CheckRateLimit(key string) bool {
	s.rateLimiter.mu.Lock()
	defer s.rateLimiter.mu.Unlock()

	limit, exists := s.rateLimiter.limits[key]
	if !exists {
		return true
	}

	// Reset counter if window has passed
	if time.Since(limit.lastReset) > time.Minute {
		limit.count = 0
		limit.lastReset = time.Now()
	}

	limit.count++
	return limit.count <= 100 // Default limit of 100 requests per minute
}

// cleanupLoop periodically cleans up old rate limit entries
func (s *Security) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.rateLimiter.cleanupCh:
			return
		case <-ticker.C:
			s.rateLimiter.mu.Lock()
			for key, limit := range s.rateLimiter.limits {
				if time.Since(limit.lastReset) > 24*time.Hour {
					delete(s.rateLimiter.limits, key)
				}
			}
			s.rateLimiter.mu.Unlock()
		}
	}
}
