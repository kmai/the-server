package tracing

import (
	"errors"
	"fmt"
)

var ErrUnsupportedExporterType = errors.New("unsupported exporter type: %s")

func UnsupportedExporterType(exporterType string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedExporterType, exporterType)
}
