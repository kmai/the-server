package routes

import (
	"fmt"
	"net/http"

	"github.com/kmai/the-server/internal/providers/logging"
)

// GetRoot returns a welcome response and nothing else.
func GetRoot(writer http.ResponseWriter, req *http.Request) {
	if _, err := writer.Write([]byte("welcome")); err != nil {
		log := logging.GetLoggerFromContext(req.Context())
		log.Error(fmt.Errorf("error while writing response: %w", err))
	}
}
