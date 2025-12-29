package server

import (
	"encoding/json"
	"time"

	"github.com/ayushanand18/crazyhttp/internal/ratelimiter"
	"github.com/ayushanand18/crazyhttp/pkg/constants"
	"github.com/ayushanand18/crazyhttp/pkg/types"
)

type method struct {
	Method constants.HttpMethodTypes
	URL    string
	s      *server

	// utility
	rateLimiter *ratelimiter.RateLimiter

	description           string
	inputSchema           interface{}
	outputSchema          interface{}
	name                  string
	handler               types.HandlerFunc
	decoder               types.HttpDecoder
	encoder               types.HttpEncoder
	beforeServeMiddleware types.HttpRequestMiddleware
	afterServeMiddleware  types.HttpResponseMiddleware
	options               types.MethodOptions
}

type Method interface {
	Serve(types.HandlerFunc) Method

	WithDecoder(decoder types.HttpDecoder) Method
	WithEncoder(encoder types.HttpEncoder) Method
	WithBeforeServe(middleware types.HttpRequestMiddleware) Method
	WithAfterServe(middleware types.HttpResponseMiddleware) Method
	WithOptions(options types.MethodOptions) Method

	WithRateLimit(types.RateLimitOptions) Method
	WithDescription(desc string) Method
	WithInputSchema(schema interface{}) Method
	WithOutputSchema(schema interface{}) Method
	WithName(name string) Method
}

func NewMethod(httpMethod constants.HttpMethodTypes, url string, s *server) Method {
	return &method{
		Method: httpMethod,
		URL:    url,
		s:      s,
	}
}

func (m *method) Serve(handler types.HandlerFunc) Method {
	m.handler = handler

	if _, ok := m.s.routeMatchMap[m.URL]; !ok {
		m.s.routeMatchMap[m.URL] = make(map[constants.HttpMethodTypes]*method)
	}

	// if the combination exists, reassign it
	m.s.routeMatchMap[m.URL][m.Method] = m

	return m
}

func DecodeJsonRequest[T any](in interface{}) (T, error) {
	var out T
	raw, err := json.Marshal(in)
	if err != nil {
		return out, err
	}

	err = json.Unmarshal(raw, &out)
	return out, err
}

func (m *method) WithDescription(desc string) Method {
	m.description = desc
	return m
}

func (m *method) WithInputSchema(schema interface{}) Method {
	m.inputSchema = schema
	return m
}

func (m *method) WithOutputSchema(schema interface{}) Method {
	m.outputSchema = schema
	return m
}

func (m *method) WithName(name string) Method {
	m.name = name
	return m
}

func (m *method) WithRateLimit(options types.RateLimitOptions) Method {
	m.rateLimiter = ratelimiter.NewRateLimiter(options.Limit, time.Duration(options.BucketDurationInSeconds)*time.Second)

	return m
}

func (m *method) WithDecoder(decoder types.HttpDecoder) Method {
	m.decoder = decoder
	return m
}

func (m *method) WithEncoder(encoder types.HttpEncoder) Method {
	m.encoder = encoder
	return m
}

func (m *method) WithBeforeServe(middleware types.HttpRequestMiddleware) Method {
	m.beforeServeMiddleware = middleware
	return m
}

func (m *method) WithAfterServe(middleware types.HttpResponseMiddleware) Method {
	m.afterServeMiddleware = middleware
	return m
}

func (m *method) WithOptions(options types.MethodOptions) Method {
	m.options = options
	return m
}
