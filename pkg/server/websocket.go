package server

import (
	"net/http"
	"time"

	"github.com/ayushanand18/crazyhttp/internal/ratelimiter"
	"github.com/ayushanand18/crazyhttp/pkg/types"
)

type websocket struct {
	Url string
	s   *server

	rateLimiter *ratelimiter.RateLimiter

	decoder               types.HttpDecoder
	encoder               types.HttpEncoder
	beforeServeMiddleware types.HttpRequestMiddleware
	afterServeMiddleware  types.HttpResponseMiddleware

	options types.WebSocketOption

	description string
	name        string
}

type WebSocket interface {
	Serve(types.WebsocketHandlerFunc)

	// Decoder for every message received
	WithDecoder(decoder types.HttpDecoder) WebSocket
	// Encoder for every message sent
	WithEncoder(encoder types.HttpEncoder) WebSocket
	// Middleware to run before every message is served
	WithBeforeServe(middleware types.HttpRequestMiddleware) WebSocket
	// Middleware to run after every message is sent
	WithAfterServe(middleware types.HttpResponseMiddleware) WebSocket
	// Name of the websocket endpoint - for Swagger API documentation
	WithName(name string) WebSocket
	// Description of the websocket endpoint - for Swagger API documentation
	WithDescription(desc string) WebSocket
	// WithOptions to add serve options
	WithOptions(options types.WebSocketOption) WebSocket
	// WithRateLimiter to add rate limiting
	// rate limit will be applied on each message received
	// key in context with which rate limiting will be done can be set using RateLimitOptions.ContextKey
	WithRateLimit(options types.RateLimitOptions) WebSocket
	// HandleHandshake to handle custom handshake
	HandleHandshake(types.WebSocketHandshakeFunc) WebSocket
}

func NewWebsocket(url string, s *server) WebSocket {
	return &websocket{Url: url, s: s}
}

func (ws *websocket) Serve(handler types.WebsocketHandlerFunc) {
	fun := ws.GetWebSocketHandlerFunc(handler)
	ws.s.mux.HandleFunc(ws.Url, http.HandlerFunc(fun))
}

func (ws *websocket) WithDecoder(decoder types.HttpDecoder) WebSocket {
	ws.decoder = decoder
	return ws
}

func (ws *websocket) WithEncoder(encoder types.HttpEncoder) WebSocket {
	ws.encoder = encoder
	return ws
}

func (ws *websocket) WithBeforeServe(middleware types.HttpRequestMiddleware) WebSocket {
	ws.beforeServeMiddleware = middleware
	return ws
}

func (ws *websocket) WithAfterServe(middleware types.HttpResponseMiddleware) WebSocket {
	ws.afterServeMiddleware = middleware
	return ws
}

func (ws *websocket) WithName(name string) WebSocket {
	ws.name = name
	return ws
}

func (ws *websocket) WithDescription(desc string) WebSocket {
	ws.description = desc
	return ws
}

func (ws *websocket) HandleHandshake(fn types.WebSocketHandshakeFunc) WebSocket {
	return ws
}

func (ws *websocket) WithOptions(options types.WebSocketOption) WebSocket {
	ws.options = options
	return ws
}

func (ws *websocket) WithRateLimit(options types.RateLimitOptions) WebSocket {
	ws.rateLimiter = ratelimiter.NewRateLimiter(options.Limit, time.Duration(options.BucketDurationInSeconds)*time.Second)

	return ws
}
