package router

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/kmai/the-server/internal/models"
	"github.com/kmai/the-server/internal/providers/config"
	"github.com/kmai/the-server/internal/providers/database"
	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/kmai/the-server/internal/providers/tracing"
	"github.com/kmai/the-server/internal/router/routes"
	"github.com/kmai/the-server/internal/router/routes/user"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func Start(ctx context.Context) {
	ctx = config.LoadConfiguration(ctx)

	// Initialize Subsystems
	log := logging.GetLoggerFromContext(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tracerProvider, err := tracing.TracerInit(ctx)
	if err != nil {
		log.Panicf("there was a problem starting the tracer: %v", err)
	}

	tracerShutdown := tracing.GetTracerShutdownSignal(tracerProvider)
	defer tracerShutdown(ctx)

	tracer := otel.Tracer(viper.GetString("service"), oteltrace.WithInstrumentationVersion(contrib.Version()))

	router := chi.NewRouter()

	setupRouterMiddleware(router, tracer, log)

	dbConnection, err := database.GetDatabaseConnection(ctx)
	if err != nil {
		log.Panicf("error while creating database dbConnection(s): %v", err)
	}

	err = dbConnection.AutoMigrate(models.User{})
	if err != nil {
		log.Panicf("could not finish auto-migrating the database models: %v", err)
	}

	router.Route("/welcome", func(r chi.Router) {
		r.Get("/", routes.GetRoot)
	})

	router.Route("/users", func(r chi.Router) {
		handler := user.Handler{Database: dbConnection, Logger: log}
		r.Get("/", handler.GetUserList)
		r.Get("/{userID}", handler.GetUser)
		r.Post("/new", handler.PostUser)
	})

	server := NewServer(ctx, router)

	if err := server.ListenAndServe(); err != nil {
		log.Errorf("server encountered an error during execution: %v", err)
	}
}
