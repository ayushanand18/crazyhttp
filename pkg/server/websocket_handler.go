package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	ashttp "github.com/ayushanand18/crazyhttp/internal/http"
	"github.com/ayushanand18/crazyhttp/pkg/constants"
	"github.com/ayushanand18/crazyhttp/pkg/errors"
	"github.com/ayushanand18/crazyhttp/pkg/types"
	gws "github.com/gorilla/websocket"
)

func websocketHandler(
	ctx context.Context,
	conn *gws.Conn,
	w http.ResponseWriter,
	r *http.Request,
	ws *websocket,
	handler types.WebsocketHandlerFunc,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	requestChannel := make(chan types.WebsocketStreamChunk)
	responseChannel := make(chan types.WebsocketStreamChunk)

	// helper to close all channels once
	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			close(requestChannel)
			close(responseChannel)
		})
	}

	// attach to context
	ctx = context.WithValue(ctx, constants.WebsocketRequestChannel, requestChannel)
	ctx = context.WithValue(ctx, constants.WebsocketResponseChannel, responseChannel)

	// Reader goroutine
	go func() {
		defer func() {
			cancel()
			closeAll()
		}()

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				if gws.IsCloseError(err,
					gws.CloseGoingAway,
					gws.CloseNormalClosure) {
					slog.Info("WebSocket connection closed by client")
				} else {
					slog.Error("Error receiving WebSocket message", "error", err)
				}
				return
			}

			if ws.rateLimiter != nil {
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
				ws.rateLimiter.Allow(k)
			}

			msg, err := ashttp.GetDefaultSerialization(message)
			if err != nil {
				w.WriteHeader(errors.DecodeErrorToHttpErrorStatus(err))
				return
			}

			select {
			case requestChannel <- types.WebsocketStreamChunk{
				MessageType: types.WebsocketMessageType(mt),
				Data:        msg,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Handler goroutine
	go func() {
		defer cancel()
		handler(ctx)
	}()

	// Writer loop (main goroutine)
	for {
		select {
		case chunk, ok := <-responseChannel:
			if !ok {
				return
			}

			if chunk.MessageType == types.WebsocketUnspecifiedMessage {
				chunk.MessageType = types.WebsocketTextMessage
			}

			var headers map[string][]string
			var encoded []byte
			var err error

			if ws.encoder != nil {
				headers, encoded, err = ws.encoder(ctx, chunk.Data)
			} else {
				headers, encoded, err = ashttp.DefaultHttpEncode(ctx, chunk.Data)
			}
			if err != nil {
				w.WriteHeader(errors.DecodeErrorToHttpErrorStatus(err))
				return
			}

			for key, value := range headers {
				w.Header().Del(key)
				for _, v := range value {
					w.Header().Add(key, v)
				}
			}

			if err := conn.WriteMessage(chunk.MessageType.ToInt(), encoded); err != nil {
				slog.Error("Error sending WebSocket message", "error", err)
				return
			}

		case <-ctx.Done():
			// someone canceled (reader, handler, or connection closed)
			closeAll()
			return
		}
	}
}
