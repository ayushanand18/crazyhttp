package http

import (
	"context"
	"encoding/json"
	"net/http"
)

func DefaultHttpEncode(ctx context.Context, response interface{}) (headers map[string][]string, body []byte, err error) {
	headers = map[string][]string{
		"Content-Type": {"application/json; charset=utf-8"},
	}

	body, err = GetDefaultSerialization(response)
	if err != nil {
		return headers, body, err
	}

	return headers, body, nil
}

func DefaultHttpDecode(ctx context.Context, r *http.Request) (outgoingRequest interface{}, err error) {
	if e := json.NewDecoder(r.Body).Decode(&outgoingRequest); e != nil {
		return outgoingRequest, err
	}

	return outgoingRequest, nil
}

func GetDefaultSerialization(req interface{}) (body []byte, err error) {
	switch v := req.(type) {
	case string:
		body = []byte(v)
	case []byte:
		body = v
	default:
		body, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}
