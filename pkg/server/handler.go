package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	ashttp "github.com/ayushanand18/crazyhttp/internal/http"
	"github.com/ayushanand18/crazyhttp/pkg/constants"
	"github.com/ayushanand18/crazyhttp/pkg/types"
	"github.com/gorilla/mux"
)

func httpDefaultHandler(
	ctx context.Context,
	w http.ResponseWriter,
	handler types.HandlerFunc,
	decoder types.HttpDecoder,
	encoder types.HttpEncoder,
	r *http.Request,
	m *method) {

	var request interface{}
	var err error

	ctx, err = defaultMiddleware(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "error in default middlewares", "err:=", err)
		return
	}

	if len(m.options.AllowedOrigins) > 0 && !ashttp.IsOriginAllowed(r.Header.Get("Origin"), m.options.AllowedOrigins) {
		w.WriteHeader(http.StatusForbidden)
		slog.ErrorContext(ctx, "origin not allowed", "origin", r.Header.Get("Origin"))
		return
	}

	if decoder != nil {
		request, err = decoder(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "error in decoding request", "err:=", err)
			return
		}
	} else {
		request, err = ashttp.DefaultHttpDecode(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "error in decoding headers", "err:=", err)
			return
		}
	}

	if m.rateLimiter != nil {
		key := ctx.Value(constants.RateLimitCustomKey)
		if key == nil || key == "" {
			key = strings.Split(r.RemoteAddr, ":")[0]
		}
		_, ok := key.(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "rate limit key is not a string", "key:=", key)
			return
		}
		m.rateLimiter.Allow(key.(string))
	}

	resp, err := handler(ctx, request)
	if err != nil {
		return
	}

	var headers map[string][]string
	var body []byte
	if encoder != nil {
		headers, body, err = encoder(ctx, resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "error in encoding response", "err:=", err)
			return
		}
	} else {
		headers, body, err = ashttp.DefaultHttpEncode(ctx, resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "error in default encoding response", "err:=", err)
			return
		}
	}

	headers = ashttp.PopulateDefaultServerHeaders(ctx, r, headers)

	for key, value := range headers {
		w.Header().Del(key)
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	if body != nil {
		_, err := w.Write(body)
		if err != nil {
			panic(err)
		}
	}
}

func defaultMiddleware(ctx context.Context, r *http.Request) (outgoingContext context.Context, err error) {
	ctx = context.WithValue(ctx, constants.HttpRequestHeaders, r.Header)

	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	ctx = context.WithValue(ctx, constants.HttpRequestURLParams, params)

	ctx = context.WithValue(ctx, constants.HttpRequestPathValues, mux.Vars(r))

	return ctx, nil
}
