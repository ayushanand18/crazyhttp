package main_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/ayushanand18/crazyhttp/pkg/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	"github.com/ayushanand18/crazyhttp/pkg/types"
	"github.com/gorilla/websocket"
)

func writeTempCert() (cleanup func(), err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})
	if err := os.WriteFile("cert.pem", certPEM, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile("key.pem", keyPEM, 0600); err != nil {
		os.Remove("cert.pem")
		return nil, err
	}
	return func() {
		os.Remove("cert.pem")
		os.Remove("key.pem")
	}, nil
}

// waitForServer polls until the TCP port accepts connections or times out.
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

func TestMain(m *testing.M) {
	cleanup, err := writeTempCert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write test TLS cert: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx := context.Background()
	server := crazyserver.NewHttpServer(ctx)
	if err := server.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Initialize failed: %v\n", err)
		os.Exit(1)
	}

	server.WebSocket("/ws-test").
		WithOptions(types.WebSocketOption{AllowedOrigins: []string{"*"}}).
		Serve(func(ctx context.Context) error {
			reqCh := ctx.Value(constants.WebsocketRequestChannel).(chan types.WebsocketStreamChunk)
			respCh := ctx.Value(constants.WebsocketResponseChannel).(chan types.WebsocketStreamChunk)
			for chunk := range reqCh {
				respCh <- types.WebsocketStreamChunk{
					Data: []byte(fmt.Sprintf("Echo: %s", chunk.Data)),
				}
			}
			return nil
		})

	go func() {
		_ = server.ListenAndServe(ctx)
	}()

	if err := waitForServer("localhost:4443"); err != nil {
		fmt.Fprintf(os.Stderr, "server not ready: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestUserRoute_WebsocketRequest(t *testing.T) {
	wsURL := "ws://localhost:4443/ws-test"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close()

	msg := "hello"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	_, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	want := fmt.Sprintf("Echo: %s", msg)
	if string(p) != want {
		t.Errorf("Expected %q, got %q", want, p)
	}
}
