package post

import (
	"context"

	infraError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error"
)

type RefreshSession struct {
	RefreshToken string `json:"refreshToken" validate:"required,gte=1,lte=300"`
}

func (a RefreshSession) Validate(ctx context.Context) error {
	return infraError.Validate(ctx, a)
}
