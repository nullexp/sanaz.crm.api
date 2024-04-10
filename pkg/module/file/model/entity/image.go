package entity

import "time"

const (
	FieldImageTitle = "title"
	FieldImageId    = "id"
)

type Image struct {
	Id        string
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
