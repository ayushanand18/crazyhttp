package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ayushanand18/crazyhttp/pkg/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
)

type MyCustomResponseType struct {
	UserId  string
	Message string
}

func encoder(ctx context.Context, response interface{}, reqErr error) (headers map[string][]string, body []byte, err error) {
	resp := response.(MyCustomResponseType)
	headers = make(map[string][]string)
	headers["X-User-Id"] = []string{resp.UserId}

	bodyBytes, err := json.Marshal(resp)
	if err != nil {
		return headers, body, err
	}

	return headers, bodyBytes, nil
}

func UserIdHandler(ctx context.Context, request interface{}) (response interface{}, err error) {
	pathValues := ctx.Value(constants.HttpRequestPathValues).(map[string]string)

	return MyCustomResponseType{
		UserId:  pathValues["user_id"],
		Message: "Hello World from GET.",
	}, nil
}

// noBodyDecoder skips JSON decoding — safe for requests with no body (e.g. GET).
func noBodyDecoder(ctx context.Context, r *http.Request) (context.Context, interface{}, error) {
	return ctx, nil, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Initialize failed: %v\n", err)
		os.Exit(1)
	}

	server.GET("/users/{user_id}").
		WithDecoder(noBodyDecoder).
		WithEncoder(encoder).
		Serve(UserIdHandler)

	go func() {
		_ = server.ListenAndServe(ctx)
	}()
	time.Sleep(500 * time.Millisecond)

	os.Exit(m.Run())
}

func TestUserRoute_WithUserIdHeader(t *testing.T) {
	addr := "localhost:4445"

	client := &http.Client{}

	resp, err := client.Get(fmt.Sprintf("http://%s/users/123", addr))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	v := MyCustomResponseType{}
	if err := json.Unmarshal(body, &v); err != nil {
		t.Errorf("Could not unmarshal json")
	}
	if v.Message != "Hello World from GET." || v.UserId != "123" {
		t.Errorf("Expected Message 'Hello World from GET.' and UserId '123', got %q", body)
	}

	userID := resp.Header.Get("X-User-Id")
	if userID != "123" {
		t.Errorf("Expected X-User-Id=123, got %q", userID)
	}
}
