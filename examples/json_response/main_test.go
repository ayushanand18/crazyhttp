package main_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	crazyserver "github.com/ayushanand18/crazyhttp/pkg/server"
	qchttp3 "github.com/quic-go/quic-go/http3"
)

// noBodyDecoder skips JSON decoding — safe for requests with no body (e.g. GET).
func noBodyDecoder(ctx context.Context, r *http.Request) (context.Context, interface{}, error) {
	return ctx, nil, nil
}

type DummyResponse struct {
	Key   string `json:"key"`
	Value uint32 `json:"value"`
}

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
