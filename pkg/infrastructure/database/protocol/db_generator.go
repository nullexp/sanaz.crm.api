package protocol

import (
	"errors"
)

type TransactionFactoryGetter interface {
	GetTransactionFactory() (TransactionFactory, error)
}

// TODO: Ambiguous definition for db first strategy, func name should be changed.
type DatabaseGenerator interface {
	Generate() error
	Init() error
}
type DatabaseController interface {
	TransactionFactoryGetter
	DatabaseGenerator
}

var (
	ErrDbMigrationFailed = errors.New("database migration failed")
	ErrDbNotFound        = errors.New("database not found")
)
