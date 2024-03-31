package protocol

type TransactionFactory interface {
	New() Transaction
}

type Transaction interface {
	DataContextGetter
	Begin() error
	Rollback() error
	RollbackUnlessCommitted()
	Commit() error
}

type DataContextGetter interface {
	// GetDataContext is used when we want to access underlying database for crud
	GetDataContext() any
}
