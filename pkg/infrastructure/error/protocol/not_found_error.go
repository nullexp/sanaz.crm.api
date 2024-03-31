package protocol

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Note that we prefer duplication over dependency than shared error type here

type NotFoundError struct {
	Id string `json:"id"`

	ErrorCode ErrorCode `json:"errorCode"`

	Description string `json:"description"`

	Details []interface{} `json:"details,omitempty"`

	Type ErrorType `json:"type"`
}

func NewNotFoundError(errorCode ErrorCode, description string) NotFoundError {
	return NotFoundError{ErrorCode: errorCode, Description: description, Type: ErrorTypeNotFound}
}

func (nfa NotFoundError) WithDetail(details ...any) NotFoundError {
	nfa.Id = uuid.NewString()

	nfa.Details = details

	return nfa
}

func (nfa NotFoundError) WithIdAndDetail(id string, details ...any) NotFoundError {
	nfa.Details = details

	nfa.Id = id

	return nfa
}

func (nfa NotFoundError) Error() string {
	dt := ErrorDto(nfa)

	raw, err := json.Marshal(dt)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (nfa NotFoundError) GetErrorCode() ErrorCode {
	return nfa.ErrorCode
}

func (nfa NotFoundError) GetDescription() string {
	return nfa.Description
}
