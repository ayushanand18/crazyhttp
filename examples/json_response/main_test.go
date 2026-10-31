package main_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	qchttp3 "github.com/quic-go/quic-go/http3"
)

type DummyResponse struct {
	Key   string `json:"key"`
	Value uint32 `json:"value"`
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

	server.GET("/json").WithDecoder(noBodyDecoder).Serve(func(ctx context.Context, req interface{}) (interface{}, error) {
		return DummyResponse{Key: "test", Value: 123}, nil
	})

	go func() {
		_ = server.ListenAndServe(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	os.Exit(m.Run())
}

func TestUserRoute_NaiveJSONResponse(t *testing.T) {
	addr := "localhost:4447"

	client := &http.Client{
		Transport: &qchttp3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/json", addr))
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

	value := DummyResponse{}
	if err := json.Unmarshal(body, &value); err != nil {
		t.Errorf("Error while parsing JSON, got %q", body)
	}
}
