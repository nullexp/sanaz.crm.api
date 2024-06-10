package post

import (
	"context"

	infraError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error"
)

type Session struct {
	Username   string `json:"username" validate:"required,gte=1,lte=30"`
	Password   string `json:"password"  validate:"required,gte=1,lte=100"`
	AccessMode string `json:"accessMode"  validate:"required,gte=1,lte=50"`
}

func (a Session) Validate(ctx context.Context) error {
	return infraError.Validate(ctx, a)
}
