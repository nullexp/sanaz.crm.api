package pgsqlite

import (
	"time"

	gormh "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/gorm"
	pg "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/gorm/pg"
	dbapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/entity"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/persistence/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewImage(getter dbapi.DataContextGetter, withTestID bool) repository.Image {
	r := gormh.NewGormUuidCompoundRepository(getter, pg.NewParser(), dbapi.NewDefaultMapper[Image, entity.Image](), withTestID, misc.Id)
	return &r
}

type Image struct {
	Id        string `gorm:"type:uuid;"`
	FileName  string
	MimeType  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (entity Image) IsIdEmpty() bool {
	return entity.Id == ""
}

func (entity *Image) SetId(id string) {
	entity.Id = id
}

func (entity Image) GetCreatedAt() time.Time {
	return entity.CreatedAt
}

func (entity Image) GetUpdatedAt() *time.Time {
	if entity.CreatedAt.Equal(entity.UpdatedAt) {
		return nil
	}

	return &entity.UpdatedAt
}

func (entity Image) GetUuid() string {
	return entity.Id
}

func (u *Image) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}
	return nil
}
