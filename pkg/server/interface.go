package server

import (
	"context"
	"net/http"

	"github.com/ayushanand18/crazyhttp/internal/constants"
	"github.com/ayushanand18/crazyhttp/internal/utils"
	"github.com/gorilla/mux"
	"github.com/quic-go/quic-go"
	qchttp3 "github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/qlog"
)

type server struct {
	// HTTP server assets
	h3server       qchttp3.Server
	mux            *mux.Router
	routeMatchMap  map[string]map[constants.HttpMethodTypes]*method
	http1ServerTLS http.Server
	http1Server    http.Server
}

type HttpServer interface {
	Initialize(context.Context) error
	ListenAndServe(context.Context) error

	// HTTP Methods
	GET(string) Method
	POST(string) Method
	PUT(string) Method
	PATCH(string) Method
	DELETE(string) Method
	HEAD(string) Method
	OPTIONS(string) Method
	CONNECT(string) Method
	TRACE(string) Method

	// Websocket
	WebSocket(string) WebSocket
}

func NewHttpServer(ctx context.Context) HttpServer {
	quicConfig := &quic.Config{
		Tracer:          qlog.DefaultConnectionTracer,
		Allow0RTT:       true,
		EnableDatagrams: true,
	}
	return &server{
		h3server: qchttp3.Server{
			Addr:            utils.GetListeningAddress(ctx),
			Handler:         nil,
			EnableDatagrams: true,
			QUICConfig:      quicConfig,
		},
		http1Server: http.Server{
			Addr: utils.GetHttp1ListeningAddress(ctx),
		},
		http1ServerTLS: http.Server{
			Addr: utils.GetHttp1TLSListeningAddress(ctx),
		},
		mux:           mux.NewRouter(),
		routeMatchMap: make(map[string]map[constants.HttpMethodTypes]*method),
	}
}
