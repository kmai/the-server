package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestID generates a UUIDv4 and injects it as a header for the request.
func RequestID(next http.Handler) http.Handler {
	interceptingFn := func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get(middleware.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		if writer.Header().Get(middleware.RequestIDHeader) == "" {
			writer.Header().Add(
				middleware.RequestIDHeader,
				requestID,
			)
		}

		// We inject the Request ID to the request context.
		ctx := context.WithValue(request.Context(), middleware.RequestIDKey, requestID)

		// Let's pass it on to the next function
		next.ServeHTTP(writer, request.WithContext(ctx))
	}

	return http.HandlerFunc(interceptingFn)
}
