package service

import (
	utility "github.com/nullexp/sanaz.crm.api/internal/module/auth/utility"
	httpapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	api "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
)

type JwtAuthenticator struct {
}

func NewJwtAuthenticator() httpapi.Authenticator {
	return JwtAuthenticator{}
}

func (JwtAuthenticator) GetModel(token string) (api.JwtClaim, error) {
	return utility.GetToken(token)
}

func (JwtAuthenticator) CheckToken(token string) (out bool, outE error) {
	return utility.CheckToken(token)
}
