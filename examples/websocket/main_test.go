package main_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/ayushanand18/crazyhttp/internal/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	"github.com/ayushanand18/crazyhttp/pkg/types"
	"github.com/gorilla/websocket"
)

// waitForServer waits until TCP port is accepting connections
func waitForServer(addr string) error {
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server not ready on %s", addr)
}

func TestUserRoute_WebsocketRequest(t *testing.T) {
	ctx := context.Background()
	addr := "localhost:4431"

	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// WebSocket endpoint
	server.WebSocket("/ws-test").
		WithOptions(types.WebSocketOption{AllowedOrigins: []string{"*"}}).
		Serve(func(ctx context.Context) error {
			reqCh := ctx.Value(constants.WebsocketRequestChannel).(chan types.WebsocketStreamChunk)
			respCh := ctx.Value(constants.WebsocketResponseChannel).(chan types.WebsocketStreamChunk)

			// Keep the handler alive to echo messages
			for chunk := range reqCh {
				respCh <- types.WebsocketStreamChunk{
					Data: []byte(fmt.Sprintf("Echo: %s", chunk.Data)),
				}
			}
			return nil
		})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(ctx); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	// Wait for server to be ready
	if err := waitForServer(addr); err != nil {
		t.Fatalf("Server not ready: %v", err)
	}

	// Connect WebSocket client
	wsURL := fmt.Sprintf("ws://%s/ws-test", addr)
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close()

	// Send a message
	msg := "hello"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	// Read the echo
	_, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	want := fmt.Sprintf("Echo: %s", msg)
	if string(p) != want {
		t.Errorf("Expected %q, got %q", want, p)
	}
}
