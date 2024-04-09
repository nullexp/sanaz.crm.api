package main

import (
	"git.omidgolestani.ir/clinic/crm.api/configs"
	"git.omidgolestani.ir/clinic/crm.api/internal/factory"
	authApplication "git.omidgolestani.ir/clinic/crm.api/internal/module/auth/application/service"
	filetApplication "git.omidgolestani.ir/clinic/crm.api/internal/module/file/application/service"
	assetEntities "git.omidgolestani.ir/clinic/crm.api/internal/module/file/persistence/repository/pgsqlite"
	filePresentation "git.omidgolestani.ir/clinic/crm.api/internal/module/file/presentation"

	dbProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model/openapi"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/log"
	assetError "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/error"
)

func initializeApi(conf configs.Config) {
	api := factory.NewApi(factory.Gin)

	db := factory.NewDatabaseController(conf.DataStorage, []dbProtocol.EntityBased{
		assetEntities.Asset{},
	}, conf.DataStorageName)

	err := db.Generate()
	if err != nil {
		log.Error.Fatalln(err)
	}
	assetRepo := factory.NewAssetRepository(factory.Data, false)
	fileStorage := factory.NewFileStorage(factory.Memory, conf.FileStorageName)

	// Initialize Modules
	subjectParser := authApplication.NewSubjectParser()
	assetApplicationService := filetApplication.NewAsset(filetApplication.AssetParam{
		AssetRepoFactory:   assetRepo,
		TransactionFactory: db,
		FileStorage:        fileStorage,
	})

	asset := filePresentation.NewAsset(assetApplicationService, subjectParser)

	api.AppendModule(asset)
	api.SetContact(openapi.Contact{Name: "Hope Golestany", Email: "hopegolestany@gmail.com", URL: "https://omidgolestani.ir"})
	api.SetInfo(openapi.Info{Version: "0.1", Description: "Api definition for clinic", Title: "Clinic Api Definition"})
	api.SetLogPolicy(model.LogPolicy{LogBody: false, LogEnabled: false})
	api.SetCors([]string{"http://localhost:8080"})
	err = api.EnableOpenApi("/openapi")
	if err != nil {
		log.Error.Fatalln(err)
	}
	api.SetErrors([]string{string(assetError.AssetNotFoundKey)})
	err = api.Run("localhost", uint(8080), "debug")
	if err != nil {
		log.Error.Fatalln(err)
	}
}
