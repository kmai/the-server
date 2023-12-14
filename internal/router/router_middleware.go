package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kmai/the-server/internal/middleware"
	"github.com/sirupsen/logrus"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func setupRouterMiddleware(router *chi.Mux, tracer oteltrace.Tracer, log *logrus.Logger) {
	// a /status endpoint for load balancers to check for health
	router.Use(chiMiddleware.Heartbeat("/status"))
	// CORS support
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		OptionsPassthrough: false,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:     []string{"Link"},
		AllowCredentials:   false,
		MaxAge:             300, // Maximum value not ignored by any of major browsers
	}))

	// Middleware that fiddles with headers and context
	router.Use(middleware.RequestID, middleware.Tracing(tracer), chiMiddleware.RealIP)
	// Let's avoid panicking if possible and return a 5xx instead.
	router.Use(chiMiddleware.Recoverer)
	// We use a logger based off sirupsen/logrus with a specific category for router log messages
	router.Use(middleware.LogrusLogger("router", log))
	// Set a timeout value on the request context (ctx), that will signal through ctx.Done() that the request has timed
	// out and further processing should be stopped.
	router.Use(chiMiddleware.Timeout(2 * time.Second))
	router.Use(chiMiddleware.RealIP)
}
