//nolint:all
package mongo

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type TransactionFactory struct {
	db *MongoDatabase
}

type Transaction struct {
	db *mongo.Database

	session mongo.Session

	committed bool
}

func NewTransactionFactory(db *MongoDatabase) protocol.TransactionFactory {
	t := TransactionFactory{db: db}

	return t
}

func (t TransactionFactory) New() protocol.Transaction {
	return &Transaction{db: t.db.Database}
}

// GetDataContext is used when we want to access underlying database for crud

func (t *Transaction) GetDataContext() any {
	return t.db
}

func (t *Transaction) Begin() error {
	session, err := t.db.Client().StartSession()
	if err != nil {
		return err
	}

	t.session = session

	err = t.session.StartTransaction(options.Transaction().SetWriteConcern(writeconcern.New(writeconcern.WMajority())))
	if err != nil {

		t.session.EndSession(context.Background())

		return err

	}

	return nil
}

func (t *Transaction) Rollback() error {
	if t.session == nil {
		return nil
	}

	err := t.session.AbortTransaction(context.Background())
	if err != nil {

		t.session.EndSession(context.Background())

		return err

	}

	t.session.EndSession(context.Background())

	return nil
}

func (t *Transaction) RollbackUnlessCommitted() {
	if !t.committed {
		_ = t.Rollback()
	}
}

func (t *Transaction) Commit() error {
	if t.session == nil {
		return nil
	}

	err := t.session.CommitTransaction(context.Background())
	if err != nil {

		err = t.session.AbortTransaction(context.Background())
		if err != nil {
			return err
		}

		// TODO: check if this method should be called when error occurred or not

		t.session.EndSession(context.Background())

		return err

	}

	t.session.EndSession(context.Background())

	t.committed = true

	return nil
}
