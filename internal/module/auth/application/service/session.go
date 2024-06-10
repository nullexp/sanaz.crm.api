package service

import (
	"context"
	"time"

	"github.com/nullexp/sanaz.crm.api/internal/module/auth/utility"
	pkservice "github.com/nullexp/sanaz.crm.api/pkg/module/auth/utility"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	dbapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	api "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	application "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application"
	request "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/request"
	response "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/response"
	userRepo "github.com/nullexp/sanaz.crm.api/pkg/module/user/persistence/repository"
)

type Session struct {
	userRepoFactory    userRepo.UserRepoFactory
	TransactionFactory protocol.TransactionFactoryGetter
	policy             application.TokenPolicy
	passwordHasher     misc.Password
}

func NewSession(ufactory userRepo.UsernameGetterFactory, tFactory dbapi.DatabaseGetter, policy application.TokenPolicy, hasher misc.Password) application.Session {
	return &Session{userRepoFactory: ufactory, dbGetter: tFactory, policy: policy, passwordHasher: hasher}
}

func (s Session) Authenticate(ctx context.Context, dto request.Session) (u response.Token, err error) {

	factory, err := s.dbGetter.GetDatabase(dto.AccessMode)
	if err != nil {
		return response.Token{}, misc.WrapUserOperationError(api.ErrorUnknownAccessMode)
	}
	tx := factory.New()
	err = tx.Begin()
	if err != nil {
		return
	}
	defer tx.RollbackUnlessCommitted()

	return s.authenticate(dto, tx)
}

func (s Session) authenticate(ctx context.Context, dto request.Session, tx dbapi.Transaction) (response.Token, error) {

	out := response.Token{}

	userRepo := s.userRepoFactory.NewUsernameGetter(tx)
	user, err := userRepo.GetByUsername(dto.Username)
	found := user.GetID() != 0

	if err != nil {
		return out, err
	}

	if !found {
		return out, misc.WrapUserOperationError(pkservice.ErrorAuthentiationError)
	}

	if user.Password != s.passwordHasher.HashAndSalt(dto.Password) {
		return out, misc.WrapUserOperationError(pkservice.ErrorAuthentiationError)
	}

	if !user.IsActive {
		return out, misc.WrapUserOperationError(pkservice.ErrorUserInactive)
	}

	if err != nil {
		return out, err
	}
	sub := utility.NewSubject(dto.AccessMode, user.ID, 0)
	accessExpireTime := time.Now().Add(s.policy.GetAccessTokenLivingDuration())
	refreshExpireTime := time.Now().Add(s.policy.GetRefreshTokenLivingDuration())

	access, err := utility.CreateToken(sub, accessExpireTime)
	if err != nil {
		return out, err
	}

	refresh, err := utility.CreateToken(sub, refreshExpireTime)
	if err != nil {
		return out, err
	}

	out.AccessToken = access
	out.RefreshToken = refresh
	out.AccessTokenExpireTime = accessExpireTime
	out.RefreshTokenExpireTime = refreshExpireTime

	return out, nil
}

func (s Session) RefreshToken(ctx context.Context, rs request.RefreshSession) (response.AccessToken, error) {

	clm, err := utility.GetToken(rs.RefreshToken)
	if err != nil {
		return response.AccessToken{}, misc.WrapUserOperationError(pkservice.ErrorInvalidToken)
	}

	return s.refreshToken(clm.Subject, rs)
}

func (s Session) refreshToken(ctx context.Context, sub string, rs request.RefreshSession) (response.AccessToken, error) {

	out := response.AccessToken{}

	accessExpireTime := time.Now().Add(s.policy.GetAccessTokenLivingDuration())
	access, err := utility.CreateTokenWithText(sub, accessExpireTime)
	if err != nil {
		return out, err
	}
	out.AccessToken = access
	out.AccessTokenExpireTime = accessExpireTime
	return out, nil
}
