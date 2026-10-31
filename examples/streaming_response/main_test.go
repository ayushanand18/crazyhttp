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
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ayushanand18/crazyhttp/pkg/constants"
	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	"github.com/ayushanand18/crazyhttp/pkg/types"
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

func HelloWorldStreaming(ctx context.Context, request interface{}) (response interface{}, err error) {
	for i := range 5 {
		time.Sleep(time.Duration(1) * time.Second)
		channel := ctx.Value(constants.StreamingResponseChannelContextKey).(chan types.StreamChunk)
		channel <- types.StreamChunk{
			Id:   uint32(i),
			Data: []byte(fmt.Sprintf("Chunk: %d \n\n", i)),
		}
	}
	return nil, nil
}

func TestMain(m *testing.M) {
	cleanup, err := writeTempCert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write test TLS cert: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx := context.Background()
	s := crazyserver.NewHttpServer(ctx)
	if err := s.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server initialization failed: %v\n", err)
		os.Exit(1)
	}

	s.GET("/streaming").
		WithOptions(types.MethodOptions{IsStreamingResponse: true}).
		Serve(HelloWorldStreaming)

	go func() {
		_ = s.ListenAndServe(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	os.Exit(m.Run())
}

func TestHTTP3Server_BasicStreamingResponse(t *testing.T) {
	addr := "localhost:4449"

	client := &http.Client{}

	resp, err := client.Get(fmt.Sprintf("http://%s/streaming", addr))
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	expected := "Chunk: 0 \n\nChunk: 1 \n\nChunk: 2 \n\nChunk: 3 \n\nChunk: 4 \n\n"
	if strings.ReplaceAll(string(body), "\r", "") != expected {
		t.Fatalf("expected streaming body:\n%q\ngot:\n%q", expected, string(body))
	}
}
