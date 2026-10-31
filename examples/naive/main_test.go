package main_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
)

const testAddr = "localhost:4441"

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

	server.GET("/test").
		WithDecoder(noBodyDecoder).
		Serve(func(ctx context.Context, request interface{}) (interface{}, error) {
			return "Hello World from GET.", nil
		})

	server.POST("/test").WithDecoder(noBodyDecoder).Serve(func(ctx context.Context, request interface{}) (interface{}, error) {
		return "Hello World from POST.", nil
	})

	go func() {
		_ = server.ListenAndServe(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	os.Exit(m.Run())
}

func TestUserRoute_NaiveGETRequest(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/test", testAddr))
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
	if string(body) != "Hello World from GET." {
		t.Errorf("Expected body \"Hello World from GET.\" got %q", body)
	}
}

func TestUserRoute_NaivePOSTRequest(t *testing.T) {
	resp, err := http.Post(fmt.Sprintf("http://%s/test", testAddr), "text/plain; utf-8", nil)
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
	if string(body) != "Hello World from POST." {
		t.Errorf("Expected body 'Hello World from POST.', got %q", body)
	}
}
