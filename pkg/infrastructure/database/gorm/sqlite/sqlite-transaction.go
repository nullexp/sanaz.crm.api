package sqlite

import (
	"sync"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"gorm.io/gorm"
)

type TransactionFactory struct {
	db *SqliteDatabase
}

type Transaction struct {
	db *gorm.DB

	tx *gorm.DB

	commited bool

	lock *sync.Mutex
}

func NewTransactionFactory(db *SqliteDatabase) protocol.TransactionFactory {
	t := TransactionFactory{db: db}

	return t
}

func (t TransactionFactory) New() protocol.Transaction {
	return &Transaction{db: t.db.Database, lock: t.db.InnerMux}
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
	t.lock.Lock()

	t.tx = t.db.Begin()

	err := t.tx.Error

	return err
}

func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return nil
	}

	err := t.tx.Rollback().Error

	defer t.lock.Unlock()

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

	t.lock.Unlock()

	return err
}
