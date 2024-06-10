package service

import (
	"context"
	"time"

	"github.com/nullexp/sanaz.crm.api/internal/module/auth/utility"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	dbapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol/specification"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	api "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	application "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application"
	request "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/request"
	response "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/response"
	authError "github.com/nullexp/sanaz.crm.api/pkg/module/auth/model/error"
	userRepo "github.com/nullexp/sanaz.crm.api/pkg/module/user/persistence/repository"
)

type SessionParam struct {
	UserRepoFactory    userRepo.UserRepoFactory
	TransactionFactory protocol.TransactionFactoryGetter
	Policy             application.TokenPolicy
	PasswordHasher     misc.Password
}

type Session struct {
	SessionParam
}

func NewSession(param SessionParam) application.Session {
	return &Session{param}
}

func (s Session) Authenticate(ctx context.Context, dto request.Session) (out response.Token, err error) {

	factory, err := s.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	out, err = s.authenticate(ctx, tx, dto)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (s Session) authenticate(ctx context.Context, tx dbapi.Transaction, dto request.Session) (response.Token, error) {

	out := response.Token{}

	userRepo := s.UserRepoFactory.NewUser(tx)
	user, err := userRepo.GetSingle(ctx, specification.NewQuerySpecification("username", api.QueryOperatorEqual, misc.NewOperand(dto.Username)))
	found := user.GetUuid() != ""

	if err != nil {
		return out, err
	}

	if !found {
		return out, authError.ErrInvalidAuth
	}

	if user.Password != s.PasswordHasher.HashAndSalt(dto.Password) {
		return out, authError.ErrInvalidAuth
	}

	if err != nil {
		return out, err
	}

	sub := utility.NewSubject(user.GetUuid(), "")
	accessExpireTime := time.Now().Add(s.Policy.GetAccessTokenLivingDuration())
	refreshExpireTime := time.Now().Add(s.Policy.GetRefreshTokenLivingDuration())

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

func (s Session) RefreshToken(ctx context.Context, rs request.RefreshSession) (out response.AccessToken, err error) {

	clm, err := utility.GetToken(rs.RefreshToken)
	if err != nil {
		return response.AccessToken{}, authError.ErrInvalidToken
	}

	accessExpireTime := time.Now().Add(s.Policy.GetAccessTokenLivingDuration())
	access, err := utility.CreateTokenWithText(clm.Subject, accessExpireTime)
	if err != nil {
		return out, err
	}
	out.AccessToken = access
	out.AccessTokenExpireTime = accessExpireTime
	return out, nil
}
