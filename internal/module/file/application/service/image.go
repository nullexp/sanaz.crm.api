package service

import (
	"context"
	"io"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/internal/module/file/application/cast"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	infraError "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
	fileProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/request"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/response"
	appService "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/service"
	imageError "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/error"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/persistence/repository"
)

type ImageParam struct {
	ImageRepoFactory   repository.ImageRepoFactory
	TransactionFactory protocol.TransactionFactoryGetter
	ImageStorage       fileProtocol.ImageStorage
}

type image struct {
	ImageParam
}

func NewImage(param ImageParam) appService.Image {
	return image{param}
}

func (a image) UploadImage(ctx context.Context, image fileProtocol.File) (out response.Image, err error) {
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

	out, err = a.uploadImage(ctx, tx, image)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (a image) uploadImage(ctx context.Context, tx protocol.Transaction, file fileProtocol.File) (out response.Image, err error) {
	repo := a.ImageRepoFactory.NewImage(tx)
	newEntity := cast.ToImageEntity(file.GetFilename(), file.GetMimeType())
	err = repo.Insert(ctx, &newEntity)
	if err != nil {
		return
	}
	out.Id = newEntity.Id

	err = a.ImageStorage.Store(file, out.Id)

	return
}

func (a image) DownloadImage(ctx context.Context, request request.Image) (out fileProtocol.File, err error) {
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

	out, err = a.downloadImage(ctx, tx, request)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (a image) downloadImage(ctx context.Context, tx protocol.Transaction, request request.Image) (out fileProtocol.File, err error) {
	if err = request.Validate(ctx); err != nil {
		return
	}

	repo := a.ImageRepoFactory.NewImage(tx)

	wantedFile, err := repo.GetById(ctx, request.Id)
	if err != nil {
		return
	}
	if wantedFile.IsIdEmpty() {
		err = imageError.ErrFileNotFound
		return
	}

	var rc io.ReadCloser
	var changeTime time.Time

	if request.Thumbnail {
		rc, changeTime, err = a.ImageStorage.RetrieveThumbnail(wantedFile.Id, request.Width)
	} else {
		rc, changeTime, err = a.ImageStorage.Retrieve(wantedFile.Id)
	}

	if err != nil {
		err = infraError.NewManagedSystemError(err, imageError.ErrFileNotFound.Id)
		return
	}
	out = fileProtocol.NewFile(rc, wantedFile.FileName, wantedFile.MimeType, changeTime)
	return
}
