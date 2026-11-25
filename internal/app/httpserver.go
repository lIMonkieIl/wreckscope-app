package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"wreckscope/internal/racefeed"
)

func (a *App) startHTTP(ctx context.Context) {
	web, err := fs.Sub(a.assets, "ui/dist")
	if err != nil {
		log.Printf("LAN assets: %v", err)
		return
	}

	bindHost := "127.0.0.1"
	if a.lan {
		bindHost = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", bindHost, a.httpPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("LAN dashboard bind failed: %v", err)
		return
	}

	mux := http.NewServeMux()
	var streamMu sync.Mutex
	streamClients := map[chan struct{}]struct{}{}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		frame := time.NewTicker(33 * time.Millisecond)
		refresh := time.NewTicker(100 * time.Millisecond)
		defer frame.Stop()
		defer refresh.Stop()
		dirty := true
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.signal:
				dirty = true
			case <-refresh.C:
				dirty = true
			case <-frame.C:
				if !dirty {
					continue
				}
				dirty = false
				streamMu.Lock()
				for ch := range streamClients {
					select {
					case ch <- struct{}{}:
					default:
					}
				}
				streamMu.Unlock()
			}
		}
	}()

	mux.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowPeer(w, r) {
			return
		}
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		ch := make(chan struct{}, 1)
		streamMu.Lock()
		streamClients[ch] = struct{}{}
		streamMu.Unlock()
		defer func() {
			streamMu.Lock()
			delete(streamClients, ch)
			streamMu.Unlock()
		}()
		send := func() bool {
			b, err := json.Marshal(a.engine.Snapshot())
			if err != nil {
				return false
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
				return false
			}
			fl.Flush()
			return true
		}
		if !send() {
			return
		}
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ctx.Done():
				return
			case <-ch:
				if !send() {
					return
				}
			}
		}
	})

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowPeer(w, r) {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(a.engine.Snapshot())
	})

	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if !a.allowPeer(w, r) {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			_ = json.NewEncoder(w).Encode(a.GetSettings())
		case http.MethodPost:
			if !a.allowControl(w, r) {
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 4096)
			defer r.Body.Close()
			var settings racefeed.Settings
			if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			saved, err := a.SaveSettings(settings)
			if err != nil {
				log.Printf("Save settings: %v", err)
				http.Error(w, "Could not save settings to disk.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(saved)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/appearance", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowControl(w, r) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		defer r.Body.Close()
		var appearance struct {
			Theme string `json:"theme"`
			Mode  string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&appearance); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		saved, err := a.SaveAppearance(appearance.Theme, appearance.Mode)
		if err != nil {
			log.Printf("Save appearance: %v", err)
			http.Error(w, "Could not save settings to disk.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(saved)
	})

	mux.HandleFunc("/api/mode", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowControl(w, r) {
			return
		}
		var v struct {
			Mode string `json:"mode"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil || (v.Mode != "demo" && v.Mode != "live") {
			http.Error(w, "invalid mode", http.StatusBadRequest)
			return
		}
		if err := a.SetMode(v.Mode); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/api/medic", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowControl(w, r) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		defer r.Body.Close()
		var v struct {
			Slot int `json:"slot"`
		}
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if v.Slot < 0 || v.Slot >= 36 {
			http.Error(w, "invalid slot", http.StatusBadRequest)
			return
		}
		medics := a.ToggleMedic(v.Slot)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"medics": medics})
	})

	mux.HandleFunc("/api/access", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowPeer(w, r) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(a.PhoneAccess())
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})

	fileServer := http.FileServer(http.FS(web))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !a.allowPeer(w, r) {
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(web, path); err != nil {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{Addr: addr, Handler: mux}
	a.httpServer = server
	log.Printf("LAN dashboard on %s (interactive)", addr)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Printf("LAN dashboard stopped: %v", err)
	}
}

func (a *App) peerAllowed(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	return a.lan && ip.IsPrivate()
}

func (a *App) allowPeer(w http.ResponseWriter, r *http.Request) bool {
	if a.peerAllowed(r) {
		return true
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}

func (a *App) allowControl(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return false
	}
	if !a.peerAllowed(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	host := r.Host
	if address, _, err := net.SplitHostPort(host); err == nil {
		host = address
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	localHost := strings.EqualFold(host, "localhost") ||
		(ip != nil && (ip.IsLoopback() || (a.lan && ip.IsPrivate())))
	origin := r.Header.Get("Origin")
	if !localHost || (origin != "" && origin != "http://"+r.Host) {
		http.Error(w, "Open the dashboard using its local address to change settings", http.StatusForbidden)
		return false
	}
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return false
	}
	return true
}
