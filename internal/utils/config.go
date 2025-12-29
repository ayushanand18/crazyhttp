package utils

import (
	"context"
	"fmt"

	"github.com/ayushanand18/crazyhttp/internal/config"
	"github.com/ayushanand18/crazyhttp/pkg/constants"
)

func GetListeningAddress(ctx context.Context) string {
	ipAddress := config.GetString(ctx, "service.http.h3.address.ip", constants.DEFAULT_SERVER_IP_ADDRESS)
	port := config.GetInt(ctx, "service.http.h3.address.port", constants.DEFAULT_SERVER_PORT)

	return fmt.Sprintf("%s:%d", ipAddress, port)
}

func GetHttp1ListeningAddress(ctx context.Context) string {
	ipAddress := config.GetString(ctx, "service.http.h1.address.ip", constants.DEFAULT_H1_SERVER_IP_ADDRESS)
	port := config.GetInt(ctx, "service.http.h1.address.port", constants.DEFAULT_H1_SERVER_PORT)

	return fmt.Sprintf("%s:%d", ipAddress, port)
}

func GetHttp1TLSListeningAddress(ctx context.Context) string {
	ipAddress := config.GetString(ctx, "service.http.h1_ssl.address.ip", constants.DEFAULT_H1_SERVER_IP_ADDRESS)
	port := config.GetInt(ctx, "service.http.h1_ssl.address.port", constants.DEFAULT_H1_SERVER_PORT)

	return fmt.Sprintf("%s:%d", ipAddress, port)
}

func GetMcpListeningAddress(ctx context.Context) string {
	ipAddress := config.GetString(ctx, "service.mcp.address.ip", constants.DEFAULT_MCP_SERVER_IP_ADDRESS)
	port := config.GetInt(ctx, "service.mcp.address.port", constants.DEFAULT_MCP_SERVER_PORT)

	return fmt.Sprintf("%s:%d", ipAddress, port)
}
