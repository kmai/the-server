package database

import (
	"errors"
	"fmt"
)

var ErrUnsupportedDatabaseEngine = errors.New("unsupported database engine")

func UnsupportedDatabaseEngine(engine string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedDatabaseEngine, engine)
}

var ErrGenericDatabaseConnection = errors.New("generic database connection error")

func GenericDatabaseConnection(err string) error {
	return fmt.Errorf("%w: %s", ErrGenericDatabaseConnection, err)
}
