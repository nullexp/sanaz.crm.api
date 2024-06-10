package pgsqlite

import (
	"time"

	"github.com/google/uuid"
	gormh "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm"
	pg "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm/pg"
	dbapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/model/entity"
	"github.com/nullexp/sanaz.crm.api/pkg/module/file/persistence/repository"
	"gorm.io/gorm"
)

func NewAsset(getter dbapi.DataContextGetter, withTestID bool) repository.Asset {
	r := gormh.NewGormUuidCompoundRepository(getter, pg.NewParser(), dbapi.NewDefaultMapper[Asset, entity.Asset](), withTestID, misc.Id)
	return &r
}

type Asset struct {
	Id        string `gorm:"type:uuid;"`
	FileName  string
	MimeType  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (entity Asset) IsIdEmpty() bool {
	return entity.Id == ""
}

func (entity *Asset) SetId(id string) {
	entity.Id = id
}

func (entity Asset) GetCreatedAt() time.Time {
	return entity.CreatedAt
}

func (entity Asset) GetUpdatedAt() *time.Time {
	if entity.CreatedAt.Equal(entity.UpdatedAt) {
		return nil
	}

	return &entity.UpdatedAt
}

func (entity Asset) GetUuid() string {
	return entity.Id
}

func (u *Asset) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}
	return nil
}
