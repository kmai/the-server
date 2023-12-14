package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/kmai/the-server/internal/providers/database/mysql"
	"github.com/kmai/the-server/internal/providers/database/sqlite"
	"github.com/kmai/the-server/internal/providers/logging"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func GetDatabaseConnection(ctx context.Context) (*gorm.DB, error) {
	log := logging.GetLoggerFromContext(ctx)
	config := &gorm.Config{
		Logger: NewLogger(ctx),
	}

	engine := strings.ToLower(viper.GetString("database.engine"))
	log.Debugf("getting database connection for engine %s", engine)

	var databaseConnection *gorm.DB

	var err error

	switch engine {
	case "sqlite":
		databaseConnection, err = sqlite.GetDatabaseConnection(config)

	case "mysql-simple":
		databaseConnection, err = mysql.GetSimpleDatabaseConnection(ctx, config)

	case "mysql-split":
		databaseConnection, err = mysql.GetSplitDatabaseConnection(ctx, config)
	default:
		return nil, UnsupportedDatabaseEngine(engine)
	}

	if err != nil {
		return nil, GenericDatabaseConnection(
			fmt.Sprintf("error obtaining database connection of type %s: %v", engine, databaseConnection),
		)
	}

	return databaseConnection, nil
}
