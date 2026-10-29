package mux1

import (
	"net/http"
	"path"
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
	cleanPath := path.Clean(r.URL.Path)

	m.mu.RLock()
	methods, pathExists := m.routes[cleanPath]
	m.mu.RUnlock()

	if !pathExists {
		if fallbackPath, ok := strings.CutSuffix(cleanPath, "/"); ok {
			if _, fallbackExists := m.routes[fallbackPath]; fallbackExists {
				target := fallbackPath
				// Preserve query string during redirect if one exists
				if r.URL.RawQuery != "" {
					target += "?" + r.URL.RawQuery
				}
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}
		}
		http.NotFound(w, r)
		return
	}

	handler, methodExists := methods[r.Method]
	if !methodExists {
		var allowed []string
		for allowedMethod := range methods {
			allowed = append(allowed, allowedMethod)
		}
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handler(w, r)
}

func (m *MyMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "GET"
	route := pattern
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		method = strings.ToUpper(parts[0])
		route = parts[1]
	}

	if m.routes[route] == nil {
		m.routes[route] = make(map[string]http.HandlerFunc)
	}

	m.routes[route][method] = handler
}
