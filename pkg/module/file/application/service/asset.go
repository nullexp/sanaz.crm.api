package service

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/request"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/response"
)

type Asset interface {
	UploadAsset(ctx context.Context, asset protocol.File) (response.Asset, error)
	DownloadAsset(ctx context.Context, id request.AssetId) (protocol.File, error)
}
