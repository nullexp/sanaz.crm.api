package protocol

import (
	"time"
)

const (
	FieldCreatedAt = "created_at"
	FieldUpdatedAt = "updated_at"
)

// Entity Every entities embed this entity
type Entity struct {
	Id        any
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (e Entity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e Entity) GetUpdatedAt() *time.Time {
	return e.UpdatedAt
}

func (e Entity) IsIdEmpty() bool {
	val, ok := e.Id.(string)

	if ok && val == "" {
		return true
	}

	valNumber, ok := e.Id.(int64)

	if ok && valNumber == 0 {
		return true
	}

	return false
}

func (e *Entity) SetId(id any) {
	e.Id = id
}

// EntityBased handle shared operation over entities.
type (
	EntityBased interface {
		GetCreatedAt() time.Time
		GetUpdatedAt() *time.Time
		Identity
	}
	Identity interface {
		IsIdEmpty() bool
	}

	UuIdGetter interface {
		GetUuid() string
	}
	IdGetter interface {
		GetId() int64
	}

	UuIdIdentity interface {
		Identity
		UuIdGetter
	}

	IdIdentity interface {
		Identity
		IdGetter
	}
)
