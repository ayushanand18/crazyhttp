package http

import (
	"context"
	"net/http"
)

func PopulateDefaultServerHeaders(ctx context.Context, r *http.Request, headers map[string][]string) map[string][]string {
	if headers == nil {
		headers = make(map[string][]string)
	}

	headers["X-Server"] = []string{"crazyhttp"}
	// relay the origin back since we check for allowed origins, earlier
	headers["Access-Control-Allow-Origin"] = []string{r.Header.Get("Origin")}
	headers["Access-Control-Allow-Methods"] = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"}
	headers["Access-Control-Allow-Headers"] = []string{"Content-Type", "Authorization"}
	headers["Access-Control-Allow-Credentials"] = []string{"true"}
	headers["Access-Control-Max-Age"] = []string{"86400"}

	return headers
}
