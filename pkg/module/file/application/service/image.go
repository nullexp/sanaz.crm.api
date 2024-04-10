package service

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/request"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/response"
)

type Image interface {
	UploadImage(ctx context.Context, asset protocol.File) (response.Image, error)
	DownloadImage(ctx context.Context, id request.Image) (protocol.File, error)
}
