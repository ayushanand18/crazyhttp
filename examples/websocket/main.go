package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ayushanand18/crazyhttp/pkg/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	"github.com/ayushanand18/crazyhttp/pkg/types"
)

func main() {
	ctx := context.Background()

	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		log.Fatalf("Server failed to Initialize: %v", err)
	}

	server.WebSocket("/ws-test").
		WithOptions(types.WebSocketOption{
			AllowedOrigins: []string{"*"},
		}).
		Serve(func(ctx context.Context) error {
			reqChanel := ctx.Value(constants.WebsocketRequestChannel).(chan types.WebsocketStreamChunk)
			respChanel := ctx.Value(constants.WebsocketResponseChannel).(chan types.WebsocketStreamChunk)

			for chunk := range reqChanel {
				fmt.Printf("Received chunk: ID=%d, Type=%d, Data=%s\n", chunk.Id, chunk.MessageType, string(chunk.Data))
				respChanel <- types.WebsocketStreamChunk{
					Data: []byte(fmt.Sprintf("Echo: %s", string(chunk.Data))),
				}
			}

			return nil
		})

	if err := server.ListenAndServe(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
