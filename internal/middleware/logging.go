package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// LogrusLogger is a middleware that will produce request log events for sirupsen/logrus.
func LogrusLogger(category string, logger logrus.FieldLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		interceptingFn := func(writer http.ResponseWriter, request *http.Request) {
			span := oteltrace.SpanContextFromContext(request.Context())
			reqID := middleware.GetReqID(request.Context())
			wrapResponseWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			startTime := time.Now()

			// This is deferred to be called only after the whole chain of Handlers has run.
			defer func() {
				remoteIP, _, err := net.SplitHostPort(request.RemoteAddr)
				if err != nil {
					remoteIP = request.RemoteAddr
				}

				scheme := "http"

				if request.TLS != nil {
					scheme = "https"
				}

				fields := logrus.Fields{
					"status_code":      wrapResponseWriter.Status(),
					"bytes":            wrapResponseWriter.BytesWritten(),
					"duration":         int64(time.Since(startTime)),
					"duration_display": time.Since(startTime).String(),
					"category":         category,
					"remote_ip":        remoteIP,
					"proto":            request.Proto,
					"scheme":           scheme,
					"method":           request.Method,
					"host":             request.Host,
					"uri":              request.RequestURI,
				}
				if len(reqID) > 0 {
					fields["request_id"] = reqID
				}

				// Do we have a valid span/trace context?
				if span.IsValid() {
					fields["trace_id"] = span.TraceID().String()
					fields["span"] = span.SpanID().String()
				}

				logger.WithFields(fields).Info("Inbound Request")
			}()

			next.ServeHTTP(wrapResponseWriter, request)
		}

		return http.HandlerFunc(interceptingFn)
	}
}
