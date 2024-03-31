package protocol

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Note that we prefer duplication over dependency than shared error type here

type UserOperationError struct {
	Id string `json:"id"`

	ErrorCode ErrorCode `json:"errorCode"`

	Description string `json:"description"`

	Details []interface{} `json:"details,omitempty"`

	Type ErrorType `json:"type"`
}

func NewUserOperationError(errorCode ErrorCode, description string) UserOperationError {
	return UserOperationError{ErrorCode: errorCode, Description: description, Type: ErrorTypeUserOperation}
}

func (uoe UserOperationError) WithDetail(details ...any) UserOperationError {
	uoe.Id = uuid.NewString()

	uoe.Details = details

	return uoe
}

func (uoe UserOperationError) WithIdAndDetail(id string, details ...any) UserOperationError {
	uoe.Details = details

	uoe.Id = id

	return uoe
}

func (uoe UserOperationError) Error() string {
	dt := ErrorDto(uoe)

	raw, err := json.Marshal(dt)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (uoe UserOperationError) GetErrorCode() ErrorCode {
	return uoe.ErrorCode
}

func (uoe UserOperationError) GetDescription() string {
	return uoe.Description
}
