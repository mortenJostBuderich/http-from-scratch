package mux1

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type MyMux struct {
	routes map[string]map[string]http.HandlerFunc
	mu     sync.RWMutex
}

func NewMyMux() *MyMux {
	return &MyMux{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

func (m *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	m.mu.RLock()
	defer m.mu.RUnlock()

	methods, pathExists := m.routes[path]

	if !pathExists {
		var fallbackPath string
		if before, ok := strings.CutSuffix(path, "/"); ok {
			fallbackPath = before
		} else {
			fallbackPath = path + "/"
		}

		// If the alternate path exists, redirect the client
		if _, fallbackExists := m.routes[fallbackPath]; fallbackExists {
			target := fallbackPath
			// Preserve query string during redirect if one exists
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		http.NotFound(w, r)
		return
	}

	handler, methodExists := methods[r.Method]
	if !methodExists {
		// Collect all allowed methods for this path to build the Allow header
		var allowed []string
		for allowedMethod := range methods {
			allowed = append(allowed, allowedMethod)
		}
		allowed = append(allowed, "OPTIONS") // Standard best practice

		w.Header().Set("Allow", strings.Join(allowed, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 Method Not Allowed\n")
		return
	}

	handler(w, r)
}

func (m *MyMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "GET"
	path := pattern
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		method = strings.ToUpper(parts[0])
		path = parts[1]
	}

	if m.routes[path] == nil {
		m.routes[path] = make(map[string]http.HandlerFunc)
	}

	m.routes[path][method] = handler
}
