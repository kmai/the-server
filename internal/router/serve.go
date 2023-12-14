package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/spf13/viper"
)

func NewServer(_ context.Context, handler http.Handler) *http.Server {
	log := logging.Init()
	portNumber := viper.GetInt("server.port")
	log.WithField("category", "startup").Infof("starting server in port %d", portNumber)

	return &http.Server{
		Handler:           handler,
		Addr:              fmt.Sprintf(":%d", portNumber),
		ReadHeaderTimeout: 2 * time.Second,
	}
}
