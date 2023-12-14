package sqlite

import (
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func isForeignKeysEnabled() bool {
	return viper.GetBool("database.sqlite.enableForeignKeys")
}

func GetDatabaseConnection(config *gorm.Config) (*gorm.DB, error) {
	var dsn string

	switch strings.ToLower(viper.GetString("database.sqlite.storage")) {
	case "in-memory":
		dsn = ":memory:"
	case "file":
		dsn = viper.GetString("database.sqlite.file.path")
	}

	if isForeignKeysEnabled() {
		dsn += "?_pragma=foreign_keys(1)"
	}

	databaseConnection, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return nil, DatabaseConnectionError(err.Error())
	}

	return databaseConnection, nil
}
