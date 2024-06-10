package cast

import (
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/model/entity"
)

func ToImageEntity(fileName, mimeType string) entity.Image {
	return entity.Image{
		FileName: fileName,
		MimeType: mimeType,
	}
}
