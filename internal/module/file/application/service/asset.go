package service

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/internal/module/file/application/cast"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	infraError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error/protocol"
	fileProtocol "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/file/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/request"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/response"
	appService "github.com/nullexp/sanaz.crm.api/pkg/module/file/application/service"
	assetError "github.com/nullexp/sanaz.crm.api/pkg/module/file/model/error"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/persistence/repository"
)

type AssetParam struct {
	AssetRepoFactory   repository.AssetRepoFactory
	TransactionFactory protocol.TransactionFactoryGetter
	FileStorage        fileProtocol.FileStorage
}

type asset struct {
	AssetParam
}

func NewAsset(param AssetParam) appService.Asset {
	return asset{param}
}

func (a asset) UploadAsset(ctx context.Context, asset fileProtocol.File) (out response.Asset, err error) {
	factory, err := a.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	out, err = a.uploadAsset(ctx, tx, asset)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (a asset) uploadAsset(ctx context.Context, tx protocol.Transaction, file fileProtocol.File) (out response.Asset, err error) {
	repo := a.AssetRepoFactory.NewAsset(tx)
	newEntity := cast.ToAssetEntity(file.GetFilename(), file.GetMimeType())
	err = repo.Insert(ctx, &newEntity)
	if err != nil {
		return
	}
	out.Id = newEntity.Id

	err = a.FileStorage.Store(file, out.Id)

	return
}

func (a asset) DownloadAsset(ctx context.Context, assetId request.AssetId) (out fileProtocol.File, err error) {
	factory, err := a.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	out, err = a.downloadAsset(ctx, tx, assetId)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (a asset) downloadAsset(ctx context.Context, tx protocol.Transaction, assetId request.AssetId) (out fileProtocol.File, err error) {
	if err = assetId.Validate(ctx); err != nil {
		return
	}

	repo := a.AssetRepoFactory.NewAsset(tx)

	wantedFile, err := repo.GetById(ctx, assetId.Id)
	if err != nil {
		return
	}
	if wantedFile.IsIdEmpty() {
		err = assetError.ErrFileNotFound
		return
	}

	rc, changeTime, err := a.FileStorage.Retrieve(wantedFile.Id)
	if err != nil {
		err = infraError.NewManagedSystemError(err, assetError.ErrFileNotFound.Id)
		return
	}

	out = fileProtocol.NewFile(rc, wantedFile.FileName, wantedFile.MimeType, changeTime)

	return
}
