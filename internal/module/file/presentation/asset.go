package presentation

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"

	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/request"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/application/dto/response"
	application "github.com/nullexp/sanaz.crm.api/pkg/module/file/application/service"

	httpapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/multipart"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/openapi"
)

const AssetBaseURL = "/assets"

func NewAsset(app application.Asset, parser misc.SubjectParser) httpapi.Module {
	return Asset{AssetService: app, parser: parser}
}

type Asset struct {
	AssetService application.Asset
	parser       misc.SubjectParser
}

func (u Asset) GetRequestHandlers() []*httpapi.RequestDefinition {
	return []*httpapi.RequestDefinition{
		u.Post(),
		u.Get(),
	}
}

func (u Asset) GetBaseURL() string {
	return AssetBaseURL
}

func (u Asset) Post() *httpapi.RequestDefinition {
	return &httpapi.RequestDefinition{
		Route:       "",
		Method:      http.MethodPost,
		Summary:     "Upload a file or any asset for storing in the system",
		Description: "This api is used to store or upload a file",
		OperationId: "postFile",

		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Description: "The id of given file",
				Status:      http.StatusCreated,
				Dto:         &response.Asset{Id: uuid.NewString()},
			},
		},
		FileParts: []httpapi.MultipartFileDefinition{&multipart.FileDefinition{Name: httpapi.KeyFile, Single: true, Optional: false, MinSize: 1, MaxSize: 25 * misc.MB}},
		Handler: func(req httpapi.Request) {
			fileHeader, _ := req.GetFile(httpapi.KeyFile)

			fl, err := fileHeader.OpenFile()
			if err != nil {
				req.ReturnStatus(http.StatusBadRequest, err)
				return
			}
			res, err := u.AssetService.UploadAsset(context.Background(), fl)
			req.Negotiate(http.StatusCreated, err, res)
		},
	}
}

func (u Asset) Get() *httpapi.RequestDefinition {
	idDef := misc.NewQueryDefinition(misc.Id,
		[]misc.QueryOperator{
			misc.QueryOperatorContain, misc.QueryOperatorEqual,
			misc.QueryOperatorNotContain, misc.QueryOperatorNotEqual,
		},
		misc.DataTypeString)

	params := []httpapi.RequestParameter{
		{Definition: idDef, Query: false, Optional: false},
	}

	return &httpapi.RequestDefinition{
		Route:       "/:id",
		Parameters:  params,
		Method:      http.MethodGet,
		Summary:     "Download a file which is stored in the system",
		Description: "This api is used to get a file by id",
		OperationId: "getFile",

		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Description: "Given file",
				Status:      http.StatusOK,
				IsFile:      true,
			},
		},
		Handler: func(req httpapi.Request) {
			id := req.MustGet(idDef.GetName()).(string)
			file, err := u.AssetService.DownloadAsset(context.Background(), request.AssetId{Id: id})
			req.WriteFile(http.StatusOK, err, file)
		},
	}
}

const (
	AssetManagement  = "Asset Management"
	AssetDescription = "Upload and download files"
)

func (a Asset) GetTag() openapi.Tag {
	return openapi.Tag{
		Name:        AssetManagement,
		Description: AssetDescription,
	}
}
