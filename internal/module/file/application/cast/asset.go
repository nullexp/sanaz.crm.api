package cast

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/entity"
)

func ToAssetEntity(fileName, mimeType string) entity.Asset {
	return entity.Asset{
		FileName: fileName,
		MimeType: mimeType,
	}
}
