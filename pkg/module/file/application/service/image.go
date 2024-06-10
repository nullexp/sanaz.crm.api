package service

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/file/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/request"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/response"
)

type Image interface {
	UploadImage(ctx context.Context, asset protocol.File) (response.Image, error)
	DownloadImage(ctx context.Context, id request.Image) (protocol.File, error)
}
