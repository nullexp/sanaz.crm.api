package utility

import (
	"os"
	"syscall"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"github.com/spf13/afero"
)

func NormalizeError(err error) error {
	casted, ok := err.(*os.PathError)
	if ok {
		if casted.Err == afero.ErrFileNotFound || casted.Err == afero.ErrFileExists || casted.Err == os.ErrNotExist {
			return protocol.ErrFileNotExist
		}
		_, ok = casted.Err.(syscall.Errno)
		if ok {
			return protocol.ErrFileNotExist
		}
	}

	return err
}
