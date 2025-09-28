package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ayushanand18/crazyhttp/internal/config"
	"github.com/ayushanand18/crazyhttp/internal/constants"
	"github.com/ayushanand18/crazyhttp/internal/tls"
	"github.com/ayushanand18/crazyhttp/internal/utils"
)

func (s *server) Initialize(ctx context.Context) error {
	if config.GetBool(ctx, "service.tls.generate_if_missing", true) && checkIfTlsCertificateIsMissing(ctx) {
		if err := tls.GenerateSelfSignedCert(ctx); err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %v", err)
		}
	}

	tlsConfig := tls.GenerateTLSConfig(ctx)
	root := &rootHandler{mux: s.mux, s: s}

	s.h3server.Handler = root
	s.h3server.TLSConfig = tlsConfig
	s.h3server.TLSConfig.NextProtos = []string{"h3"}

	h1Handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if on H/1 advertise H/3
		w.Header().Set("Alt-Svc", fmt.Sprintf(`h3=":%s"; ma=2592000,h3-29=":%s"; ma=2592000`, s.h3server.Addr[strings.LastIndex(s.h3server.Addr, ":")+1:], s.h3server.Addr[strings.LastIndex(s.h3server.Addr, ":")+1:]))
		root.ServeHTTP(w, r)
	})
	s.http1Server.Handler = h1Handler
	s.http1ServerTLS.Handler = h1Handler
	s.http1ServerTLS.TLSConfig = tlsConfig

	return nil
}

func (s *server) ListenAndServe(ctx context.Context) error {
	utils.PrintStartBanner()

	// populate mux from routeMatchMap
	for pattern, methods := range s.routeMatchMap {
		for httpMethod, m := range methods {
			s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				// call the right handler (streaming or normal)
				if m.options.IsStreamingResponse {
					streamingDefaultHandler(r.Context(), w, m.handler, m.decoder, m.encoder, r, m)
				} else {
					httpDefaultHandler(r.Context(), w, m.handler, m.decoder, m.encoder, r, m)
				}
			}).Methods(string(httpMethod))
		}
	}

	// wire mux as the Handler for all servers
	s.h3server.Handler = s.mux
	s.http1Server.Handler = s.mux
	s.http1ServerTLS.Handler = s.mux

	errChan := make(chan error, 3)

	if config.GetBool(ctx, "service.http.h3.enabled", false) {
		go func() {
			slog.InfoContext(ctx, "Starting HTTP/3 server", "port", s.h3server.Addr)
			errChan <- s.h3server.ListenAndServe()
		}()
	}

	if config.GetBool(ctx, "service.http.h1.enabled", false) {
		go func() {
			slog.InfoContext(ctx, "Starting HTTP/1.1 + Alt-Svc server", "port", s.http1Server.Addr)
			errChan <- s.http1Server.ListenAndServe()
		}()
	}

	if config.GetBool(ctx, "service.http.h1_ssl.enabled", false) {
		go func() {
			slog.InfoContext(ctx, "Starting HTTPS server", "port", s.http1ServerTLS.Addr)
			errChan <- s.http1ServerTLS.ListenAndServeTLS("", "")
		}()
	}

	return <-errChan
}

func (s *server) GET(url string) Method {
	return NewMethod(constants.HttpMethodGet, url, s)
}

func (s *server) POST(url string) Method {
	return NewMethod(constants.HttpMethodPost, url, s)

}
func (s *server) PUT(url string) Method {
	return NewMethod(constants.HttpMethodPut, url, s)
}

func (s *server) PATCH(url string) Method {
	return NewMethod(constants.HttpMethodPatch, url, s)
}

func (s *server) DELETE(url string) Method {
	return NewMethod(constants.HttpMethodDelete, url, s)
}

func (s *server) HEAD(url string) Method {
	return NewMethod(constants.HttpMethodHead, url, s)
}

func (s *server) OPTIONS(url string) Method {
	return NewMethod(constants.HttpMethodOptions, url, s)
}

func (s *server) CONNECT(url string) Method {
	return NewMethod(constants.HttpMethodConnect, url, s)
}

func (s *server) TRACE(url string) Method {
	return NewMethod(constants.HttpMethodTrace, url, s)
}

func (s *server) WebSocket(url string) WebSocket {
	return NewWebsocket(url, s)
}

// serve the HTTP request, and provide a response
func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			slog.ErrorContext(r.Context(), "panic recovered: %v\n%s", err, debug.Stack())
		}
	}()
	for k, v := range injectConstantHeaders() {
		w.Header().Set(k, v)
	}
	DumpRequest(r)

	h.mux.ServeHTTP(w, r)
}
