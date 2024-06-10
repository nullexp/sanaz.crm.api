package main

import (
	"github.com/nullexp/sanaz.crm.api/configs"
	"github.com/nullexp/sanaz.crm.api/internal/factory"
	authApplication "github.com/nullexp/sanaz.crm.api/internal/module/auth/application/service"
	filetApplication "github.com/nullexp/sanaz.crm.api/internal/module/file/application/service"
	assetEntities "github.com/nullexp/sanaz.crm.api/internal/module/file/persistence/repository/pgsqlite"
	filePresentation "github.com/nullexp/sanaz.crm.api/internal/module/file/presentation"

	authPresentation "github.com/nullexp/sanaz.crm.api/internal/module/auth/presentation"

	dbProtocol "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/openapi"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/log"
	authError "github.com/nullexp/sanaz.crm.api/pkg/module/auth/model/error"
	assetError "github.com/nullexp/sanaz.crm.api/pkg/module/file/model/error"
)

func initializeApi(conf configs.Config) {
	api := factory.NewApi(factory.Gin)

	db := factory.NewDatabaseController(conf.DataStorage, []dbProtocol.EntityBased{
		assetEntities.Asset{},
		assetEntities.Image{},
	}, conf.DataStorageName)

	err := db.Generate()
	if err != nil {
		log.Error.Fatalln(err)
	}
	assetRepo := factory.NewAssetRepository(factory.Data, false)
	imageRepo := factory.NewImageRepository(factory.Data, false)

	fileStorage := factory.NewFileStorage(factory.Memory, conf.FileStorageName)
	imageStorage := factory.NewImageStorage(factory.Memory, conf.FileStorageName)

	// Initialize Modules
	subjectParser := authApplication.NewSubjectParser()
	assetApplicationService := filetApplication.NewAsset(filetApplication.AssetParam{
		AssetRepoFactory:   assetRepo,
		TransactionFactory: db,
		FileStorage:        fileStorage,
	})

	imageApplicationService := filetApplication.NewImage(filetApplication.ImageParam{
		ImageRepoFactory:   imageRepo,
		TransactionFactory: db,
		ImageStorage:       imageStorage,
	})

	asset := filePresentation.NewAsset(assetApplicationService, subjectParser)
	image := filePresentation.NewImage(imageApplicationService, subjectParser)

	auth := authPresentation.NewSession(nil)

	api.AppendModule(asset)
	api.AppendModule(image)
	api.AppendModule(auth)

	api.SetContact(openapi.Contact{Name: "Hope Golestany", Email: "hopegolestany@gmail.com", URL: "https://omidgolestani.ir"})
	api.SetInfo(openapi.Info{Version: "0.1", Description: "Api definition for clinic", Title: "Clinic Api Definition"})
	api.SetLogPolicy(model.LogPolicy{LogBody: false, LogEnabled: false})
	api.SetCors([]string{"http://localhost:8080"})
	err = api.EnableOpenApi("/openapi")
	if err != nil {
		log.Error.Fatalln(err)
	}
	api.SetErrors([]string{string(assetError.AssetNotFoundKey), string(authError.AuthNotFoundKey), string(authError.AuthInvalidTokenKey)})
	err = api.Run("localhost", uint(8080), "debug")
	if err != nil {
		log.Error.Fatalln(err)
	}
}
