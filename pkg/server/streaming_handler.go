package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/ayushanand18/crazyhttp/internal/constants"
	ashttp "github.com/ayushanand18/crazyhttp/internal/http"
	"github.com/ayushanand18/crazyhttp/pkg/errors"
	"github.com/ayushanand18/crazyhttp/pkg/types"
)

func streamingDefaultHandler(
	ctx context.Context,
	w http.ResponseWriter,
	handler types.HandlerFunc,
	decoder types.HttpDecoder,
	encoder types.HttpEncoder,
	r *http.Request,
	m *method,
) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	if len(m.options.AllowedOrigins) > 0 &&
		!ashttp.IsOriginAllowed(r.Header.Get("Origin"), m.options.AllowedOrigins) {
		w.WriteHeader(http.StatusForbidden)
		slog.ErrorContext(ctx, "origin not allowed", "origin", r.Header.Get("Origin"))
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by this server!", http.StatusInternalServerError)
		return
	}

	// master cancel context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan types.StreamChunk)

	// safe channel close
	var once sync.Once
	closeAll := func() {
		once.Do(func() { close(ch) })
	}

	// expose channel in ctx
	ctx = context.WithValue(ctx, constants.StreamingResponseChannelContextKey, ch)

	// worker goroutine: decode, rate-limit, call handler
	go func() {
		defer func() {
			cancel()
			closeAll()
		}()

		var request interface{}
		var err error

		if decoder != nil {
			request, err = decoder(ctx, r)
			if err != nil {
				w.WriteHeader(errors.DecodeErrorToHttpErrorStatus(err))
				return
			}
		}

		if m.rateLimiter != nil {
			key := ctx.Value(constants.RateLimitCustomKey)
			if key == nil || key == "" {
				key = strings.Split(r.RemoteAddr, ":")[0]
			}
			k, ok := key.(string)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				slog.ErrorContext(ctx, "rate limit key is not a string", "key:=", key)
				return
			}
			m.rateLimiter.Allow(k)
		}

		_, err = handler(ctx, request)
		if err != nil {
			w.WriteHeader(errors.DecodeErrorToHttpErrorStatus(err))
			return
		}
	}()

	// writer loop (main goroutine)
	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return
			}

			var headers map[string][]string
			var encoded []byte
			var err error

			if encoder != nil {
				headers, encoded, err = encoder(ctx, chunk.Data)
			} else {
				headers, encoded, err = ashttp.DefaultHttpEncode(ctx, chunk.Data)
			}
			if err != nil {
				w.WriteHeader(errors.DecodeErrorToHttpErrorStatus(err))
				cancel()
				return
			}

			for key, value := range headers {
				w.Header().Del(key)
				for _, v := range value {
					w.Header().Add(key, v)
				}
			}

			if _, err := w.Write(encoded); err != nil {
				cancel()
				return
			}

			flusher.Flush()

		case <-ctx.Done():
			closeAll()
			return
		}
	}
}
