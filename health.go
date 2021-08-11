// Package health implements a health check as the logical summation of boolean
// components.
package health

import (
	"context"
	"net/http"
	"sync"
)

type contextKeyType int

const contextKey contextKeyType = iota + 1

// NewContext returns a copy of ctx which carries check.
func NewContext(ctx context.Context, check *Check) context.Context {
	return context.WithValue(ctx, contextKey, check)
}

// FromContext returns the Check the given Context carries.
//
// FromContext panics in case the given Context carries no Check.
func FromContext(ctx context.Context) *Check {
	return ctx.Value(contextKey).(*Check)
}

// Check implements a health check as the logical summation of named boolean
// components.
//
// Instances of Check implement http.Handler and their functionality is safe
// for concurrent use by multiple callers.
type Check struct {
	mu         sync.RWMutex
	components map[string]struct{}
}

// Set sets the given components of the Check to not failing.
func (c *Check) Set(components ...string) {
	if len(components) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, component := range components {
		delete(c.components, component)
	}
}

// Unset sets the given components of the Check to failing.
func (c *Check) Unset(components ...string) {
	if len(components) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.components == nil {
		c.components = make(map[string]struct{}, len(components))
	}

	for _, component := range components {
		c.components[component] = struct{}{}
	}
}

// Healthy reports the logical summation the Check's components.
//
// Healthy reports true for any Check on which no component has been Set or
// Unset.
func (c *Check) Healthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.components) == 0
}

// Failing appends the failing components of the Check to dst and returns the
// resulting slice.
//
// A Check on which Failing returns at least 1 component is always unhealty.
func (c *Check) Failing(dst []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for component := range c.components {
		dst = append(dst, component)
	}

	return dst
}

// ServeHTTP implements http.Handler for Check.
//
// ServeHTTP responds with http.StatusServiceUnavailable when the Check is
// unhealthy and with http.StatusNoContent (HEAD) or http.StatusOK (GET) when
// it's not.
//
// In all other cases the ServeHTTP responds with http.StatusMethodNotAllowed.
func (c *Check) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		respondWith(wr, http.StatusMethodNotAllowed)

		return
	case http.MethodGet, http.MethodHead:
		break
	}

	healthy := c.Healthy()

	if r.Method == http.MethodHead {
		if healthy {
			wr.WriteHeader(http.StatusNoContent)
		} else {
			wr.WriteHeader(http.StatusServiceUnavailable)
		}

		return
	}

	if healthy {
		respondWith(wr, http.StatusOK)

		return
	}

	respondWith(wr, http.StatusServiceUnavailable)
}

func respondWith(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
