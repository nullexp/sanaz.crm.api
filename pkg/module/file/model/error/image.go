package error

import (
	errorProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
)

const (
	ImageNotFoundKey  errorProtocol.ErrorCode = "api.image.NOT_FOUND"
	ImageNotFoundDesc string                  = "عکس مورد نظر پیدا نشد"
)

var ErrImageNotFound = errorProtocol.NewNotFoundError(ImageNotFoundKey, ImageNotFoundDesc)
