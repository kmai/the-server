package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func setConfigurationDefaults() {
	viper.SetDefault("environment", "development")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("service", "example-service")
	viper.SetDefault("telemetry.tracing.exporter", "otlp_http")
	viper.SetDefault("telemetry.tracing.processor", "batch")
	viper.SetDefault("telemetry.tracing.otlp_grpc.endpoint", "localhost:30080")
	viper.SetDefault("telemetry.tracing.otlp.tls.enabled", false)
	viper.SetDefault("telemetry.tracing.otlp_http.endpoint", "localhost:4318")
	viper.SetDefault("telemetry.tracing.otlp_http.url_path", "/v1/traces")

	viper.SetDefault("database.engine", "mysql-simple")
	viper.SetDefault("database.sqlite.storage", "in-memory")
	viper.SetDefault("database.sqlite.file.path", "application.db")
	viper.SetDefault("database.sqlite.enableForeignKeys", true)

	viper.SetDefault("database.mysql.simple.networkType", "tcp")
	viper.SetDefault("database.mysql.simple.hostname", "localhost")
	viper.SetDefault("database.mysql.simple.port", "3306")
	viper.SetDefault("database.mysql.simple.databaseName", "database")

	viper.SetDefault("database.mysql.simple.collation", "utf8mb4_general_ci")
	viper.SetDefault("database.mysql.simple.location", "UTC")
	viper.SetDefault("database.mysql.simple.params", "parseTime=true")

	viper.SetDefault("database.mysql.simple.timeouts.dial", "2s")
	viper.SetDefault("database.mysql.simple.timeouts.read", "2s")
	viper.SetDefault("database.mysql.simple.timeouts.write", "2s")

	viper.SetDefault("database.mysql.simple.tls.enabled", false)

	viper.SetDefault("database.mysql.split.master.networkType", "tcp")
	viper.SetDefault("database.mysql.split.master.hostname", "localhost")
	viper.SetDefault("database.mysql.split.master.port", "3306")
	viper.SetDefault("database.mysql.split.master.databaseName", "database")

	viper.SetDefault("database.mysql.split.replica.networkType", "tcp")
	viper.SetDefault("database.mysql.split.replica.hostname", "localhost")
	viper.SetDefault("database.mysql.split.replica.port", "3306")
	viper.SetDefault("database.mysql.split.replica.databaseName", "database")

	viper.SetDefault("database.mysql.split.collation", "utf8mb4_general_ci")
	viper.SetDefault("database.mysql.split.location", "UTC")
	viper.SetDefault("database.mysql.split.params", "parseTime=true")

	viper.SetDefault("database.mysql.split.timeouts.dial", "2s")
	viper.SetDefault("database.mysql.split.timeouts.read", "2s")
	viper.SetDefault("database.mysql.split.timeouts.write", "2s")

	viper.SetDefault("database.mysql.split.tls.enabled", false)
}

func readConfiguration() error {
	viper.SetEnvPrefix("SERVER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()
	viper.SetConfigName("config")     // name of config file (without extension)
	viper.SetConfigType("yaml")       // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/app/conf/") // path to look for the config file in
	viper.AddConfigPath(".")          // optionally look for config in the working directory

	// Let's read the config file
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil
		}

		return fmt.Errorf("error while reading configuration: %w", err)
	}

	return nil
}

func LoadConfiguration(ctx context.Context) context.Context {
	setConfigurationDefaults()

	var log *logrus.Logger

	if err := readConfiguration(); err != nil {
		log = &logrus.Logger{}
		log.Error(fmt.Errorf("error while reading configuration: %w", err))
	} else {
		log = logging.Init()
		ctx = logging.SetLoggerToContext(ctx, log)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", e.Name)
	})

	viper.WatchConfig()

	return ctx
}
