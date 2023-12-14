package tracing

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/spf13/viper"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlphttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

//nolint:ireturn,nolintlint
func getOpenTelemetryGRPCSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var transportCredentials credentials.TransportCredentials

	if viper.GetBool("telemetry.tracing.otlp_grpc.tls.enabled") {
		transportCredentials = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		})
	} else {
		transportCredentials = insecure.NewCredentials()
	}

	conn, err := grpc.DialContext(ctx, viper.GetString("telemetry.tracing.otlp_grpc.endpoint"),
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	exporter, err = otlpgrpc.New(ctx, otlpgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn,nolintlint
func getOpenTelemetryHTTPSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter

	opts := []otlphttp.Option{
		otlphttp.WithEndpoint(viper.GetString("telemetry.tracing.otlp_http.endpoint")),
		otlphttp.WithURLPath(viper.GetString("telemetry.tracing.otlp_http.url_path")),
	}

	if !viper.GetBool("telemetry.tracing.otlp_http.tls.enabled") {
		opts = append(opts, otlphttp.WithInsecure())
	}

	exporter, err := otlphttp.New(ctx,
		opts...,
	)
	if err != nil {
		return exporter, fmt.Errorf("error while creating exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn,nolintlint
func getZipkinSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	log := logging.GetLoggerFromContext(ctx)

	exporter, err := zipkin.New(viper.GetString("telemetry.tracing.zipkin.endpoint"),
		zipkin.WithClient(&http.Client{}),
		zipkin.WithLogr(logrusr.New(log).WithName("exporter")),
	)
	if err != nil {
		return exporter, fmt.Errorf("error while creating exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn,nolintlint
func getStdoutSpanExporter(_ context.Context) (trace.SpanExporter, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return exporter, fmt.Errorf("error while creating exporter: %w", err)
	}

	return exporter, nil
}
