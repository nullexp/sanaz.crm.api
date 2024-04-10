package cast

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/entity"
)

func ToImageEntity(fileName, mimeType string) entity.Image {
	return entity.Image{
		FileName: fileName,
		MimeType: mimeType,
	}
}
