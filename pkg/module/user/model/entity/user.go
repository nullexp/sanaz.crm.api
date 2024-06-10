package entity

import "time"

type User struct {
	Id        string
	FirstName string
	LastName  string
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (entity User) IsIdEmpty() bool {
	return entity.Id == ""
}

func (entity *User) SetId(id string) {
	entity.Id = id
}

func (entity User) GetCreatedAt() time.Time {
	return entity.CreatedAt
}

func (entity User) GetUpdatedAt() *time.Time {
	if entity.CreatedAt.Equal(entity.UpdatedAt) {
		return nil
	}
	return &entity.UpdatedAt
}

func (entity User) GetUuid() string {
	return entity.Id
}
