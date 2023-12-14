package tracing

import (
	"context"
	"fmt"
	"strings"

	logrusr "github.com/bombsimon/logrusr/v4"
	"github.com/kmai/the-server/internal/providers/logging"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// TracerInit returns a configured OpenTelemetry TracerProvider with a Resource, Exporter and Span Processor.
func TracerInit(ctx context.Context) (*tracesdk.TracerProvider, error) {
	log := logging.GetLoggerFromContext(ctx)

	var res *resource.Resource

	// We get the exporter from the config
	exp, err := GetSpanExporterFromConfig(ctx)
	if err != nil {
		log.Errorf("error while creating exporter: %v", err)

		return nil, err
	}

	// Then, we construct the resource
	if res, err = resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		//  resource.WithFromEnv(), resource.WithProcess(), resource.WithHost(),
		resource.WithContainer(), resource.WithContainerID(),
		// resource.WithProcessRuntimeName(), resource.WithProcessRuntimeVersion(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(viper.GetString("service")),
			semconv.DeploymentEnvironmentKey.String(viper.GetString("environment")),
		),
	); err != nil {
		return nil, fmt.Errorf("error while generating resource: %w", err)
	}

	// Afterward, we get the span processor from the config
	processor := GetSpanProcessorFromConfig(exp)

	// And, finally, we create the TracerProvider
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(processor),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1))),
	)

	otel.SetTracerProvider(tracerProvider)

	theLogger := logrusr.New(logging.GetLoggerFromContext(ctx))

	otel.SetLogger(theLogger)
	otel.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))

	return tracerProvider, nil
}

func GetTracerShutdownSignal(tp *tracesdk.TracerProvider) func(ctx context.Context) {
	return func(ctx context.Context) {
		log := logging.GetLoggerFromContext(ctx)

		err := tp.Shutdown(ctx)
		if err != nil {
			log.Errorf("error while shutting down trace provider: %v", err)
		}
	}
}

// GetSpanExporterFromConfig returns a configured exporter based on the configuration, or panics if the requested
// exporter isn't implemented.
//
//nolint:ireturn,nolintlint
func GetSpanExporterFromConfig(ctx context.Context) (tracesdk.SpanExporter, error) {
	var exporter tracesdk.SpanExporter

	var err error

	exporterType := viper.GetString("telemetry.tracing.exporter")

	switch strings.ToLower(exporterType) {
	case "otlp_grpc":
		exporter, err = getOpenTelemetryGRPCSpanExporter(ctx)
	case "otlp_http":
		exporter, err = getOpenTelemetryHTTPSpanExporter(ctx)
	case "zipkin":
		exporter, err = getZipkinSpanExporter(ctx)
	case "stdout":
		exporter, err = getStdoutSpanExporter(ctx)
	default:
		return nil, UnsupportedExporterType(exporterType)
	}

	if err != nil {
		return nil, fmt.Errorf("error while instantiating exporter: %w", err)
	}

	return exporter, nil
}

// GetSpanProcessorFromConfig returns a configured processor based on the configuration and the specified exporter,
// or panics if the requested span processor isn't implemented.
//
//nolint:ireturn,nolintlint
func GetSpanProcessorFromConfig(exporter tracesdk.SpanExporter) tracesdk.SpanProcessor {
	var processor tracesdk.SpanProcessor

	processorType := viper.GetString("telemetry.tracing.processor")

	switch strings.ToLower(processorType) {
	case "batch":
		processor = tracesdk.NewBatchSpanProcessor(exporter)
	case "simple":
		processor = tracesdk.NewSimpleSpanProcessor(exporter)
	default:
		logrus.Panicf("processor of type %s isn't implemented", processorType)
	}

	return processor
}
