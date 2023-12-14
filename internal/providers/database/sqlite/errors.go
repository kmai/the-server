package sqlite

import (
	"errors"
	"fmt"
)

var ErrSqliteDatabaseConnection = errors.New("sqlite database connection error")

func DatabaseConnectionError(err string) error {
	return fmt.Errorf("%w: %s", ErrSqliteDatabaseConnection, err)
}
