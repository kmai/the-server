package mysql

import (
	"errors"
	"fmt"
)

var ErrMySQLDatabaseConnection = errors.New("mysql database connection error")

func DatabaseConnectionError(err string) error {
	return fmt.Errorf("%w: %s", ErrMySQLDatabaseConnection, err)
}
