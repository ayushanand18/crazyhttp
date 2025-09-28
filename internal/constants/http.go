package constants

type HttpMethodTypes string

const (
	HttpMethodGet     HttpMethodTypes = "GET"
	HttpMethodPost    HttpMethodTypes = "POST"
	HttpMethodPut     HttpMethodTypes = "PUT"
	HttpMethodPatch   HttpMethodTypes = "PATCH"
	HttpMethodDelete  HttpMethodTypes = "DELETE"
	HttpMethodHead    HttpMethodTypes = "HEAD"
	HttpMethodOptions HttpMethodTypes = "OPTIONS"
	HttpMethodConnect HttpMethodTypes = "CONNECT"
	HttpMethodTrace   HttpMethodTypes = "TRACE"
)

type ResponseTypes int

const (
	ResponseTypeBaseResponse      ResponseTypes = iota
	ResponseTypeStreamingResponse ResponseTypes = 1
	ResponseTypeJSONResponse      ResponseTypes = 2
)

type ContextKeys string

const (
	StreamingResponseChannelContextKey ContextKeys = "response_channel"
	HttpRequestHeaders                 ContextKeys = "request_headers"
	HttpRequestURLParams               ContextKeys = "request_url_params"
	HttpRequestPathValues              ContextKeys = "request_path_values"
	RateLimitCustomKey                 ContextKeys = "rate_limit_custom_key"

	// websocket specific context keys
	WebsocketRequestChannel  ContextKeys = "websocket_request_channel"
	WebsocketResponseChannel ContextKeys = "websocket_response_channel"
)
