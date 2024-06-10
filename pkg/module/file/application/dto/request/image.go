package request

import (
	"context"

	infraError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error"
)

type Image struct {
	Id        string `json:"id" validate:"required,uuid"`
	Thumbnail bool   `json:"thumbnail"`
	Width     int    `json:"width"`
}

func (a Image) Validate(ctx context.Context) error {
	return infraError.Validate(ctx, a)
}
