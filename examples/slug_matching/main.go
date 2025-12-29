package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/ayushanand18/crazyhttp/pkg/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
)

type MyCustomRequestType struct {
	UserName string `json:"user_name"`
}

type MyCustomResponseType struct {
	UserId   string
	UserName string
	Message  string
}

func UserIdHandler(ctx context.Context, request interface{}) (response interface{}, err error) {
	req, err := crazyserver.DecodeJsonRequest[MyCustomRequestType](request)
	if err != nil {
		slog.ErrorContext(ctx, "failed to transform request", "err", err)
		return nil, err
	}

	pathValues := ctx.Value(constants.HttpRequestPathValues).(map[string]string)

	return &MyCustomResponseType{
		UserId:   pathValues["user_id"],
		UserName: req.UserName,
		Message:  "Hello World from GET.",
	}, nil
}

func main() {
	ctx := context.Background()

	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		log.Fatalf("Server failed to Initialize: %v", err)
	}

	server.GET("/users/{user_id}").Serve(UserIdHandler)

	if err := server.ListenAndServe(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
