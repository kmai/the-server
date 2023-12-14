package mysql

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/spf13/viper"
	driver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func getParamsFromSimpleConfig() map[string]string {
	paramsString := viper.GetString("database.mysql.simple.params")
	// The DSN standard says parameters are k=v, separated by &
	parameterList := strings.Split(paramsString, "&")

	// It's easier to make the map with the expected size
	parameters := make(map[string]string, len(parameterList))

	if len(parameterList) > 0 {
		for i := 0; i < len(parameterList); i++ {
			parameterSlice := strings.Split(parameterList[i], "=")
			// If there is no assignment, we should ignore it
			if len(parameterSlice) != 2 { //nolint:gomnd,nolintlint
				break
			}

			parameters[parameterSlice[0]] = parameterSlice[1]
		}
	}

	return parameters
}

func getLocationFromSimpleConfig() *time.Location {
	loc, err := time.LoadLocation(viper.GetString("database.mysql.simple.location"))
	if err != nil {
		return time.UTC
	}

	return loc
}

func getDialTimeoutFromSimpleConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.simple.timeouts.dial"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getReadTimeoutFromSimpleConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.simple.timeouts.read"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getWriteTimeoutFromSimpleConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.simple.timeouts.write"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getTLSFromSimpleConfig(ctx context.Context, cfg *mysql.Config) {
	if viper.GetBool("database.mysql.simple.tls.enabled") {
		cfg.TLSConfig = "custom"

		err := mysql.RegisterTLSConfig(cfg.TLSConfig, &tls.Config{
			MinVersion: tls.VersionTLS12,
		})
		if err != nil {
			log := logging.GetLoggerFromContext(ctx)
			log.Warnf("couldn't register TLS config for database connection, continuing without tls: %v", err)
		}
	}
}

func GetSimpleDatabaseConnection(ctx context.Context, config *gorm.Config) (*gorm.DB, error) {
	cfg := &mysql.Config{
		Net: viper.GetString("database.mysql.simple.networkType"),
		Addr: net.JoinHostPort(
			viper.GetString("database.mysql.simple.hostname"),
			viper.GetString("database.mysql.simple.port"),
		),
		User:         viper.GetString("database.mysql.simple.username"),
		Passwd:       viper.GetString("database.mysql.simple.password"),
		DBName:       viper.GetString("database.mysql.simple.databaseName"),
		Params:       getParamsFromSimpleConfig(),
		Collation:    viper.GetString("database.mysql.simple.collation"),
		Loc:          getLocationFromSimpleConfig(),
		Timeout:      getDialTimeoutFromSimpleConfig(),
		ReadTimeout:  getReadTimeoutFromSimpleConfig(),
		WriteTimeout: getWriteTimeoutFromSimpleConfig(),
	}

	getTLSFromSimpleConfig(ctx, cfg)

	databaseConn, err := gorm.Open(driver.Open(cfg.FormatDSN()), config)
	if err != nil {
		return nil, DatabaseConnectionError(err.Error())
	}

	return databaseConn, nil
}
