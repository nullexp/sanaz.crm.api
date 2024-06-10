package application

import (
	"context"
	"time"

	request "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/request"
	response "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/response"
)

type Session interface {
	Authenticate(context.Context, request.Session) (response.Token, error)
	RefreshToken(context.Context, request.RefreshSession) (response.AccessToken, error)
}

type TokenPolicy interface {
	GetAccessTokenLivingDuration() time.Duration
	GetRefreshTokenLivingDuration() time.Duration
}
