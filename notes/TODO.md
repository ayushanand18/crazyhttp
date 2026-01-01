## To Do
> List of tasks/plan of action.

### Feature Implementations
+ [x] naive H/1; H/2 and H/3 server; advertise alt-svc in H/2.
+ [x] slug based request handling
+ [x] requests params handling
+ [x] simple JSON responses
+ [x] streaming responses (LLM/media streaming use cases)
+ [x] custom encoders/decoder
+ [x] media/file responses
+ [x] support plain simple http (without TLS)
+ [ ] generic middlewares server-wide
+ [x] endpoint specific middlewares
+ [ ] web-security (csrf, cors)
+ [ ] auth support (basic auth; jwt auth)
+ [ ] out of the box rate limiting support (options in middleware)
+ [ ] monitoring (prometheous/otel standard API)

### Documentation
+ [ ] design document
+ [ ] tooling/better examples in readme
+ [ ] public API coverage

### Tests
+ [ ] 100% coverage for all internal functions
+ [ ] unittests for race-around
+ [ ] performance and benchmarking tests
