package scrape

import (
	"strings"
	"time"
)

type request struct {
	method  string
	url     string
	headers []string
	body    string

	followLocation bool
	timeout        time.Duration
}

// hasHeader returns true if the request
// has the provided header
func (r request) HasHeader(h string) bool {
	norm := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	for _, candidate := range r.headers {

		p := strings.SplitN(candidate, ":", 2)
		if norm(p[0]) == norm(h) {
			return true
		}
	}
	return false
}
