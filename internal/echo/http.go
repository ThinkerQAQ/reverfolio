package echo

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed web/*
var web embed.FS

func Handler(service Service, token string) http.Handler {
	// The local UI gets a short-lived process-scoped, HttpOnly session cookie.
	// Programmatic API and MCP clients still use the original bearer token.
	// The cookie never exposes the bearer token to JavaScript or page markup.
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		panic("could not initialize local UI session")
	}
	uiSession := hex.EncodeToString(seed)
	mux := http.NewServeMux()
	assets, err := fs.Sub(web, "web")
	if err != nil {
		panic("could not load embedded UI assets")
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(assets)))
	service.registerWorkflow(mux)
	service.registerMCPManagement(mux)
	write := func(w http.ResponseWriter, value any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	}
	handleError := func(w http.ResponseWriter, err error) {
		status := http.StatusInternalServerError
		message := "internal error"
		if errors.Is(err, ErrInvalid) {
			status = http.StatusBadRequest
			message = "invalid input"
		}
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
			message = "not found"
		}
		http.Error(w, message, status)
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { write(w, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "reverfolio_ui", Value: uiSession, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		data, _ := web.ReadFile("web/app.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	mux.HandleFunc("GET /api/works", func(w http.ResponseWriter, r *http.Request) {
		v, err := service.Store.Works(r.Context())
		if err != nil {
			handleError(w, err)
			return
		}
		write(w, v)
	})
	mux.HandleFunc("POST /api/works", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Title      string `json:"title"`
			Repository string `json:"repository"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&body) != nil || decoder.Decode(new(any)) != io.EOF {
			handleError(w, ErrInvalid)
			return
		}
		v, err := service.Store.CreateWork(r.Context(), strings.TrimSpace(body.Title), strings.TrimSpace(body.Repository))
		if err != nil {
			handleError(w, err)
			return
		}
		write(w, v)
	})
	parseID := func(r *http.Request) (int64, error) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			return 0, ErrInvalid
		}
		return id, nil
	}
	mux.HandleFunc("POST /api/works/{id}/collect", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			handleError(w, err)
			return
		}
		ctx, cancel := contextTimeout(r)
		defer cancel()
		v, err := service.Collect(ctx, id, r.Header.Get("Idempotency-Key"))
		if err != nil {
			handleError(w, err)
			return
		}
		write(w, v)
	})
	mux.HandleFunc("GET /api/works/{id}/snapshots", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			handleError(w, err)
			return
		}
		if _, err = service.Store.Work(r.Context(), id); err != nil {
			handleError(w, err)
			return
		}
		v, err := service.Store.Snapshots(r.Context(), id)
		if err != nil {
			handleError(w, err)
			return
		}
		write(w, v)
	})
	mux.HandleFunc("GET /api/works/{id}/changes", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			handleError(w, err)
			return
		}
		if _, err = service.Store.Work(r.Context(), id); err != nil {
			handleError(w, err)
			return
		}
		v, err := service.Store.Snapshots(r.Context(), id)
		if err != nil {
			handleError(w, err)
			return
		}
		// Failed release reads do not erase the most recent comparable observation.
		complete := []Snapshot{}
		for _, snap := range v {
			if snap.ReleasesOK {
				complete = append(complete, snap)
			}
		}
		if len(complete) < 2 {
			write(w, Interval{State: "baseline", Assets: []Change{}})
			return
		}
		write(w, Compare(complete[1], complete[0]))
	})
	mux.HandleFunc("GET /api/works/{id}/traffic/{metric}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			handleError(w, err)
			return
		}
		if _, err = service.Store.Work(r.Context(), id); err != nil {
			handleError(w, err)
			return
		}
		v, err := service.Store.Days(r.Context(), id, r.PathValue("metric"))
		if err != nil {
			handleError(w, err)
			return
		}
		write(w, map[string]any{"timezone": "UTC", "daily": v, "note": "Daily unique counts are not additive."})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' blob:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		// Require exact loopback Host even if the socket was reached through a proxy; prevent DNS rebinding.
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "127.0.0.1" && host != "localhost" && host != "::1" {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		apiRoute := strings.HasPrefix(r.URL.Path, "/api/")
		// Pin browser-extension access to Reverfolio's own signed manifest origin.
		// The extension is another local control surface, NOT a source of MCP tokens.
		// Chromium sets the Origin and Fetch Metadata headers; arbitrary website
		// scripts cannot impersonate the pinned chrome-extension:// origin.
		extensionOrigin := r.Header.Get("Origin")
		// Chromium extension pages with host permission may omit Origin on
		// GET (Sec-Fetch-Site=none). Origin alone is not authentication.
		extensionUI := apiRoute &&
			r.Header.Get("Sec-Fetch-Mode") == "cors" &&
			(extensionOrigin == TrustedReverfolioExtensionOrigin ||
				(extensionOrigin == "" && r.Header.Get("Sec-Fetch-Site") == "none"))
		presented := r.Header.Get("X-Reverfolio-Extension-Access")
		extensionCapabilityOK := extensionUI && service.ExtensionToken != "" &&
			subtle.ConstantTimeCompare([]byte(presented), []byte(service.ExtensionToken)) == 1
		if extensionUI {
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Origin", TrustedReverfolioExtensionOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Reverfolio-UI, Idempotency-Key, X-Reverfolio-Extension-Access")
			w.Header().Set("Access-Control-Max-Age", "300")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		if apiRoute || r.URL.Path == "/mcp" {
			r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
			provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			bearerOK := strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") &&
				token != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1
			cookie, cookieErr := r.Cookie("reverfolio_ui")
			cookieOK := cookieErr == nil && subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(uiSession)) == 1
			// A cookie is only accepted for the browser UI's same-origin fetches;
			// no cookie access to MCP, and unsafe requests require an explicit UI header.
			if cookieOK {
				site := r.Header.Get("Sec-Fetch-Site")
				if r.URL.Path == "/mcp" || site == "cross-site" || site == "same-site" ||
					(r.Method != http.MethodGet && r.Header.Get("X-Reverfolio-UI") != "1") {
					cookieOK = false
				}
			}
			// Extension can read the local API and perform explicitly marked UI
			// mutations. It cannot access /mcp and never receives a bearer key.
			extensionAuthorized := extensionCapabilityOK &&
				(r.Method == http.MethodGet ||
					(r.Method == http.MethodPost && r.Header.Get("X-Reverfolio-UI") == "1"))
			if !bearerOK && !cookieOK && !extensionAuthorized {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host && !extensionAuthorized {
				http.Error(w, "invalid origin", http.StatusForbidden)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func contextTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 90*time.Second)
}
