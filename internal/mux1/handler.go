package mux1

import (
	"fmt"
	"net/http"
	"sync"
)

type MyMux struct {
	handlerMap map[string]func(http.ResponseWriter, *http.Request)
	mu         sync.RWMutex
}

func NewMyMux() *MyMux {
	return &MyMux{make(map[string]func(http.ResponseWriter, *http.Request)), sync.RWMutex{}}
}

func (m *MyMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var pattern string
	if request.Method == "GET" {
		pattern = request.RequestURI
	} else {
		pattern = fmt.Sprintf("%s %s", request.Method, request.RequestURI)
	}

	m.mu.RLock()
	if f, ex := m.handlerMap[pattern]; ex {
		f(writer, request)
	} else {
		http.NotFound(writer, request)
	}
	m.mu.RUnlock()
}

func (m *MyMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlerMap[pattern] = handler
}
