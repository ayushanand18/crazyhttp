package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/ayushanand18/crazyhttp/internal/config"
	internalhttp "github.com/ayushanand18/crazyhttp/internal/http"
	"github.com/ayushanand18/crazyhttp/pkg/types"
	gws "github.com/gorilla/websocket"
)

func checkIfTlsCertificateIsMissing(ctx context.Context) bool {
	keyRawBytes := config.GetBytes(ctx, "service.tls.key.raw")
	if len(keyRawBytes) > 0 {
		return false
	}

	certRawBytes := config.GetBytes(ctx, "service.tls.certificate.raw")
	if len(certRawBytes) > 0 {
		return false
	}

	var keyFile, certFile string

	keyFile = config.GetString(ctx, "service.tls.key.path", "key.pem")
	certFile = config.GetString(ctx, "service.tls.certificate.path", "cert.pem")

	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return true
	}

	return false
}

func injectConstantHeaders() map[string]string {
	defaultHeaders := make(map[string]string)
	defaultHeaders["X-Server"] = "crazyhttp"

	return defaultHeaders
}

func CastParams(in interface{}, out interface{}) (interface{}, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DumpRequest prints the full HTTP request (headers + body if present)
func DumpRequest(req *http.Request) {
	// Read and restore the body so it can be read again
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
	}

	// Dump the full request (headers + body)
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error("Failed to dump HTTP request", "error", err)
		return
	}

	slog.Info("HTTP Request Dump", "method", req.Method, "url", req.URL.String(), "request", string(dump))

	// Restore the body again to ensure downstream handlers can read it
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}

// GetWebSocketHandlerFunc wraps a method onto websocket handler func
func (ws *websocket) GetWebSocketHandlerFunc(handler types.WebsocketHandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := gws.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				if len(ws.options.AllowedOrigins) > 0 &&
					!internalhttp.IsOriginAllowed(r.Header.Get("Origin"), ws.options.AllowedOrigins) {
					return false
				}
				return true
			},
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Error handling Upgrading websocket", "error", err)
			return
		}
		defer c.Close()

		websocketHandler(r.Context(), c, w, r, ws, handler)
	}
}
