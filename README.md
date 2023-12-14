# the-server

This small server was built using the following:

- [go-chi/chi](https://github.com/go-chi/chi): routing and middleware.
- [sirupsen/logrus](https://github.com/sirupsen/logrus): logging 
- [spf13/viper](https://github.com/spf13/viper): configuration management (file, environment variables, defaults, etc).
- [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go): for distributed tracing

It already contains certain features like:
- *Real IP*: the middleware will inspect multiple headers to determine the real IP of the client.
- *Tracing*: a span is generated as-is with the route and method, but it still doesn't inspect the response to complete
the span information (soon!). 
- CORS support: If using this server with a frontend, you can configure CORS to send cross-origin requests.
- Self-Recovery: if a method panics, the router will try to recover from the panic and return a 5xx instead of crashing.
- Timeout: the server will terminate connections that exceed certain duration
- Status endpoint: an endpoint in `/status` will reply back without any significant data, useful for health checks.
