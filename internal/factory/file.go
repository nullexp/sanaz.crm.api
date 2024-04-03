package factory

import (
	dmem "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/afero/disk"
	fmem "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/afero/memory"

	fileProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
)

const (
	Memory = "memory"
	Disk   = "disk"
)

func NewImageStorage(name string, param ...any) fileProtocol.ImageStorage {
	if name == "" {
		name = Test
	}
	switch name {
	case Test:
		fallthrough
	case Memory:
		return fmem.NewImageStorage()
	case Disk:
		return dmem.NewImageStorage()
	}
	panic(ErrNotImplemented)
}

func NewFileStorage(name string, param ...any) fileProtocol.FileStorage {
	if name == "" {
		name = Test
	}
	switch name {
	case Test:
		fallthrough
	case Memory:
		dir, ok := param[0].(string)
		if !ok {
			panic(ErrMissingParameter)
		}
		return fmem.NewFileStorage(dir)
	case Disk:
		dir, ok := param[0].(string)
		if !ok {
			panic(ErrMissingParameter)
		}
		return dmem.NewFileStorage(dir)
	}
	panic(ErrNotImplemented)
}
