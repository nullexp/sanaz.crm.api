package pgsqlite

import (
	dbapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	repo "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/persistence/repository"
)

type fileRepoFactory struct {
	testId bool
}

func NewFileRepoFactory(testId bool) fileRepoFactory {
	return fileRepoFactory{testId: testId}
}

func (f fileRepoFactory) NewAsset(getter dbapi.DataContextGetter) repo.Asset {
	return NewAsset(getter, f.testId)
}
