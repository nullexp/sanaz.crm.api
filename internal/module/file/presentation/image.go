package presentation

import (
	"context"
	"net/http"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"github.com/google/uuid"

	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/request"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/dto/response"
	application "git.omidgolestani.ir/clinic/crm.api/pkg/module/file/application/service"

	httpapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model/multipart"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model/openapi"
)

const ImageBaseURL = "/images"

func NewImage(app application.Image, parser misc.SubjectParser) httpapi.Module {
	return Image{ImageService: app, parser: parser}
}

type Image struct {
	ImageService application.Image
	parser       misc.SubjectParser
}

func (u Image) GetRequestHandlers() []*httpapi.RequestDefinition {
	return []*httpapi.RequestDefinition{
		u.Post(),
		u.Get(),
	}
}

func (u Image) GetBaseURL() string {
	return ImageBaseURL
}

func (u Image) Post() *httpapi.RequestDefinition {
	return &httpapi.RequestDefinition{
		Route:       "",
		Method:      http.MethodPost,
		Summary:     "Upload an image for storing in the system as image",
		Description: "This api is used to store or upload an image supporting thumbnail",
		OperationId: "postImage",

		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Description: "The id of given image",
				Status:      http.StatusCreated,
				Dto:         &response.Image{Id: uuid.NewString()},
			},
		},
		FileParts: []httpapi.MultipartFileDefinition{&multipart.FileDefinition{Name: httpapi.KeyFile, Single: true, Optional: false, MinSize: 1, MaxSize: 25 * misc.MB, SupportedTypes: []string{"image/jpeg",
			"image/png"}}},
		Handler: func(req httpapi.Request) {
			imageHeader, _ := req.GetFile(httpapi.KeyFile)

			fl, err := imageHeader.OpenFile()
			if err != nil {
				req.ReturnStatus(http.StatusBadRequest, err)
				return
			}
			res, err := u.ImageService.UploadImage(context.Background(), fl)
			req.Negotiate(http.StatusCreated, err, res)
		},
	}
}

const Thumbnail = "thumbnail"
const Width = "width"

func (u Image) Get() *httpapi.RequestDefinition {
	idDef := misc.NewQueryDefinition(misc.Id,
		[]misc.QueryOperator{
			misc.QueryOperatorEqual,
		},
		misc.DataTypeString)

	isThumb := misc.NewQueryDefinition(Thumbnail,
		[]misc.QueryOperator{
			misc.QueryOperatorEqual,
		},
		misc.DataTypeBoolean)

	width := misc.NewQueryDefinition(Width,
		[]misc.QueryOperator{
			misc.QueryOperatorEqual,
		},
		misc.DataTypeInteger)

	params := []httpapi.RequestParameter{
		{Definition: idDef, Query: false, Optional: false},
		{Definition: isThumb, Query: true, Optional: true},
		{Definition: width, Query: true, Optional: true},
	}

	return &httpapi.RequestDefinition{
		Route:       "/:id",
		Parameters:  params,
		Method:      http.MethodGet,
		Summary:     "Download a image which is stored in the system",
		Description: "This api is used to get a image by id",
		OperationId: "getFile",

		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Description: "Given image",
				Status:      http.StatusOK,
				IsFile:      true,
			},
		},
		Handler: func(req httpapi.Request) {
			id := req.MustGet(idDef.GetName()).(string)

			isThumbRaw, ok := req.Get(Thumbnail)
			isThumb := false
			if ok {
				isThumb = isThumbRaw.(bool)
			}

			widthRaw, ok := req.Get(Width)
			width := 0
			if ok {
				width = widthRaw.(int)
			}

			image, err := u.ImageService.DownloadImage(context.Background(), request.Image{Id: id, Thumbnail: isThumb, Width: width})
			req.WriteFile(http.StatusOK, err, image)
		},
	}
}

const (
	ImageManagement  = "Image Management"
	ImageDescription = "Upload and download images"
)

func (a Image) GetTag() openapi.Tag {
	return openapi.Tag{
		Name:        ImageManagement,
		Description: ImageDescription,
	}
}
