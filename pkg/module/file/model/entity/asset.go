package entity

import "time"

const (
	FieldAssetTitle = "title"
	FieldAssetId    = "id"
)

type Asset struct {
	Id        string
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
