package blog

import (
	"log"
	"net/http"
	"sync"
)

type HTMXReturner func(http.ResponseWriter, *http.Request) (string, error)

type Backend struct {
	cacheMutex sync.RWMutex
	cache      map[string]string
}

func (b *Backend) AddToMux(mux *http.ServeMux) {
	mux.HandleFunc("GET /editor/{page}", b.editorPage)
}

func (b *Backend) cacheMiddleware(h HTMXReturner) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.cacheMutex.RLock()
		if pag, ok := b.cache[r.URL.EscapedPath()]; ok {
			w.Write([]byte(pag))
			b.cacheMutex.RUnlock()
			return
		}
		b.cacheMutex.RUnlock()
		b.cacheMutex.Lock()
		defer b.cacheMutex.Unlock()
		res, err := h(w, r)
		if err != nil {
			log.Printf("error getting %v: %v", r.URL.EscapedPath(), err)
		} else {
			b.cache[r.URL.EscapedPath()] = res
		}
		w.Write([]byte(res))
	})
}
