package request

import (
	"context"

	infraError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error"
)

type AssetId struct {
	Id string `json:"id" validate:"required,uuid"`
}

func (a AssetId) Validate(ctx context.Context) error {
	return infraError.Validate(ctx, a)
}
