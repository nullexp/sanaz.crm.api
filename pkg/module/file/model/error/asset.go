package error

import (
	errorProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
)

const (
	AssetNotFoundKey  errorProtocol.ErrorCode = "api.asset.NOT_FOUND"
	AssetNotFoundDesc string                  = "فایل مورد نظر پیدا نشد"
)

var ErrFileNotFound = errorProtocol.NewNotFoundError(AssetNotFoundKey, AssetNotFoundDesc)
