package error

import (
	errorProtocol "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error/protocol"
)

const (
	AuthNotFoundKey  errorProtocol.ErrorCode = "api.auth.INVALID_AUTH"
	AuthNotFoundDesc string                  = "اطلاعات هویتی نادرست است"

	AuthInvalidTokenKey  errorProtocol.ErrorCode = "api.auth.INVALID_TOKEN"
	AuthInvalidTokenDesc string                  = "توکن ارسالی نادرست است"
)

var ErrInvalidAuth = errorProtocol.NewUserOperationError(AuthNotFoundKey, AuthNotFoundDesc)
var ErrInvalidToken = errorProtocol.NewUserOperationError(AuthInvalidTokenKey, AuthInvalidTokenDesc)
