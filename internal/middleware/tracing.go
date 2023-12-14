package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/semconv/v1.20.0/netconv"
	"go.opentelemetry.io/otel/trace"
)

// Tracing is a middleware that will inject a span to each request by using the global TracerProvider, and will
// afterward pass the request further down to other handlers (or the request handler itself).
func Tracing(tracer trace.Tracer) func(http.Handler) http.Handler {
	outerInterceptingFn := func(nestedNext http.Handler) http.Handler {
		innerInterceptingFn := func(writer http.ResponseWriter, request *http.Request) {
			// Filter out OPTIONS (not required as CORS is doing Passthrough.
			// This can also be used to filter out certain requests from getting traces
			/*
				if request.Method == http.MethodOptions {
					nestedNext.ServeHTTP(writer, request.WithContext(request.Context()))
				}
			*/
			propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))

			ctx := propagator.Extract(request.Context(), propagation.HeaderCarrier(request.Header))

			// We start the span
			ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", strings.ToUpper(request.Method), request.RequestURI),
				trace.WithAttributes(semconv.NetTransportTCP),
				trace.WithAttributes(netconv.Transport("tcp")),
				trace.WithAttributes(httpconv.ClientRequest(request)...),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Extract values to pass as headers
			carrier := propagation.MapCarrier{}
			propagator.Inject(ctx, carrier)

			for k, v := range carrier {
				if writer.Header().Get(k) == "" {
					writer.Header().Add(k, v)
				}
			}

			// We inject the trace data into the headers to propagate it downstream
			propagator.Inject(ctx, propagation.HeaderCarrier(request.Header))

			// We pass the request to the next fn
			nestedNext.ServeHTTP(writer, request.WithContext(ctx))
		}

		return http.HandlerFunc(innerInterceptingFn)
	}

	return outerInterceptingFn
}
