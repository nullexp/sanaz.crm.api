package pgsqlite

import (
	dbapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	repo "github.com/nullexp/sanaz.crm.api/pkg/module/file/persistence/repository"
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

func (f fileRepoFactory) NewImage(getter dbapi.DataContextGetter) repo.Image {
	return NewImage(getter, f.testId)
}
