package pg

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"gorm.io/gorm"
)

type TransactionFactory struct {
	db *PgDatabase
}

type Transaction struct {
	db *gorm.DB

	tx *gorm.DB

	commited bool
}

func NewTransactionFactory(db *PgDatabase) protocol.TransactionFactory {
	t := TransactionFactory{db: db}

	return t
}

func (t TransactionFactory) New() protocol.Transaction {
	return &Transaction{db: t.db.Database}
}

// GetDataContext is used when we want to access underlying database for crud

func (t *Transaction) GetDataContext() any {
	var db *gorm.DB

	if t.tx == nil {
		db = t.db
	} else {
		db = t.tx
	}

	return db
}

func (t *Transaction) Begin() error {
	t.tx = t.db.Begin()

	err := t.tx.Error

	return err
}

func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return nil
	}

	err := t.tx.Rollback().Error

	return err
}

func (t *Transaction) RollbackUnlessCommitted() {
	if !t.commited {
		_ = t.Rollback()
	}
}

func (t *Transaction) Commit() error {
	err := t.tx.Commit().Error

	t.commited = true

	return err
}
