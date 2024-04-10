package factory

import (
	fileFactory "git.omidgolestani.ir/clinic/crm.api/internal/module/file/persistence/repository/pgsqlite"
	repo "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/persistence/repository"
)

const (
	Data = "data"
)

func NewAssetRepository(name string, params ...any) repo.AssetRepoFactory {
	if name == "" {
		name = Test
	}

	switch name {
	case Test:
		fallthrough
	case Data:

		if len(params) == 0 {
			return fileFactory.NewFileRepoFactory(false)
		}
		val, _ := params[0].(bool)
		return fileFactory.NewFileRepoFactory(val)

	}
	panic(ErrNotImplemented)
}

func NewImageRepository(name string, params ...any) repo.ImageRepoFactory {
	if name == "" {
		name = Test
	}

	switch name {
	case Test:
		fallthrough
	case Data:

		if len(params) == 0 {
			return fileFactory.NewFileRepoFactory(false)
		}
		val, _ := params[0].(bool)
		return fileFactory.NewFileRepoFactory(val)

	}
	panic(ErrNotImplemented)
}
