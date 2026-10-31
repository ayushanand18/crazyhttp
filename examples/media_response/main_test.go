package main_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ayushanand18/crazyhttp/pkg/errors"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
)

// noBodyDecoder skips JSON decoding — safe for requests with no body (e.g. GET).
func noBodyDecoder(ctx context.Context, r *http.Request) (context.Context, interface{}, error) {
	return ctx, nil, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server initialization failed: %v\n", err)
		os.Exit(1)
	}

	server.GET("/audio").WithDecoder(noBodyDecoder).Serve(func(ctx context.Context, request interface{}) (interface{}, error) {
		audioBytes, err := os.ReadFile("complete_quest_requirement.mp3")
		if err != nil {
			return nil, errors.InternalServerError.New("Could not read audio file.")
		}
		return audioBytes, nil
	})

	server.GET("/html_file.html").WithDecoder(noBodyDecoder).Serve(func(ctx context.Context, request interface{}) (interface{}, error) {
		fileBytes, err := os.ReadFile("html_file.html")
		if err != nil {
			return nil, errors.InternalServerError.New("Could not read html file.")
		}
		return fileBytes, nil
	})

	go func() {
		_ = server.ListenAndServe(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	os.Exit(m.Run())
}

func TestAudioEndpoint(t *testing.T) {
	addr := "localhost:4451"
	resp, err := http.Get(fmt.Sprintf("http://%s/audio", addr))
	if err != nil {
		t.Fatalf("GET /audio failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	if len(body) == 0 {
		t.Errorf("Expected non-empty audio file body")
	}
}

func TestHTMLEndpoint(t *testing.T) {
	addr := "localhost:4451"
	resp, err := http.Get(fmt.Sprintf("http://%s/html_file.html", addr))
	if err != nil {
		t.Fatalf("GET /html_file.html failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	if len(body) == 0 {
		t.Errorf("Expected non-empty HTML file body")
	}
}
