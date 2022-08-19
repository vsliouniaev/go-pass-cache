package util

import (
	"net/http"
)

func NewCacheHandler(strategy CacheStrategy, handler http.Handler) http.Handler {
	return &cacheHandler{
		strategy: strategy,
		handler:  handler,
	}
}

type cacheHandler struct {
	strategy CacheStrategy
	handler  http.Handler
}

type CacheStrategy int

const (
	Forever = iota
	Never
)

func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.strategy == Forever {
		w.Header().Add("Cache-Control", "max-age=31536000")
	} else {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	}
	h.handler.ServeHTTP(w, r)
}
