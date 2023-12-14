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
	"gorm.io/plugin/dbresolver"
)

func getParamsFromSplitConfig() map[string]string {
	paramsString := viper.GetString("database.mysql.split.params")

	// The DSN standard says parameters are k=v, separated by &
	parameterList := strings.Split(paramsString, "&")

	// It's easier to make the map with the expected size
	parameters := make(map[string]string, len(parameterList))

	if len(parameterList) > 0 {
		for i := 0; i < len(parameterList); i++ {
			parameterSlice := strings.Split(parameterList[i], "=")
			// If there is no assignment, we should ignore it
			// TODO: determine if there are parameters that don't have a value
			if len(parameterSlice) != 2 { //nolint:gomnd,nolintlint
				break
			}

			parameters[parameterSlice[0]] = parameterSlice[1]
		}
	}

	return parameters
}

func getLocationFromSplitConfig() *time.Location {
	loc, err := time.LoadLocation(viper.GetString("database.mysql.split.location"))
	if err != nil {
		return time.UTC
	}

	return loc
}

func getDialTimeoutFromSplitConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.split.timeouts.dial"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getReadTimeoutFromSplitConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.split.timeouts.read"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getWriteTimeoutFromSplitConfig() time.Duration {
	timeout, err := time.ParseDuration(viper.GetString("database.mysql.split.timeouts.write"))
	if err != nil {
		return 2 * time.Second //nolint:gomnd,nolintlint
	}

	return timeout
}

func getTLSFromSplitConfig(ctx context.Context, cfg *mysql.Config) {
	if viper.GetBool("database.mysql.split.tls.enabled") {
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

func GetSplitDatabaseConnection(ctx context.Context, config *gorm.Config) (*gorm.DB, error) {
	masterConfig := &mysql.Config{
		Net: viper.GetString("database.mysql.split.master.networkType"),
		Addr: net.JoinHostPort(
			viper.GetString("database.mysql.split.master.hostname"),
			viper.GetString("database.mysql.split.master.port"),
		),
		User:         viper.GetString("database.mysql.split.master.username"),
		Passwd:       viper.GetString("database.mysql.split.master.password"),
		DBName:       viper.GetString("database.mysql.split.master.databaseName"),
		Params:       getParamsFromSplitConfig(),
		Collation:    viper.GetString("database.mysql.split.collation"),
		Loc:          getLocationFromSplitConfig(),
		Timeout:      getDialTimeoutFromSplitConfig(),
		ReadTimeout:  getReadTimeoutFromSplitConfig(),
		WriteTimeout: getWriteTimeoutFromSplitConfig(),
	}

	replicaConfig := &mysql.Config{
		Net: viper.GetString("database.mysql.split.replica.networkType"),
		Addr: net.JoinHostPort(
			viper.GetString("database.mysql.split.replica.hostname"),
			viper.GetString("database.mysql.split.replica.port"),
		),
		User:         viper.GetString("database.mysql.split.replica.username"),
		Passwd:       viper.GetString("database.mysql.split.replica.password"),
		DBName:       viper.GetString("database.mysql.split.replica.databaseName"),
		Params:       getParamsFromSplitConfig(),
		Collation:    viper.GetString("database.mysql.split.collation"),
		Loc:          getLocationFromSplitConfig(),
		Timeout:      getDialTimeoutFromSplitConfig(),
		ReadTimeout:  getReadTimeoutFromSplitConfig(),
		WriteTimeout: getWriteTimeoutFromSplitConfig(),
	}

	getTLSFromSplitConfig(ctx, masterConfig)
	getTLSFromSplitConfig(ctx, replicaConfig)

	resolver := dbresolver.Register(dbresolver.Config{
		// use `db2` as sources, `db3`, `db4` as replicas
		Sources:  []gorm.Dialector{driver.Open(masterConfig.FormatDSN())},
		Replicas: []gorm.Dialector{driver.Open(replicaConfig.FormatDSN())},
		// sources/replicas load balancing policy
		Policy: dbresolver.RandomPolicy{},
		// print sources/replicas mode in logger
		TraceResolverMode: true,
	})

	databaseConnection, err := gorm.Open(driver.Open(masterConfig.FormatDSN()), config)
	if err != nil {
		return nil, DatabaseConnectionError(err.Error())
	}

	err = databaseConnection.Use(resolver)
	if err != nil {
		return nil, DatabaseConnectionError(err.Error())
	}

	return databaseConnection, nil
}
