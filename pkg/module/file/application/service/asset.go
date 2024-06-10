package service

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/file/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/request"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/response"
)

type Asset interface {
	UploadAsset(ctx context.Context, asset protocol.File) (response.Asset, error)
	DownloadAsset(ctx context.Context, id request.AssetId) (protocol.File, error)
}
