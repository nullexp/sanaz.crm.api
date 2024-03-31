package protocol

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Note that we prefer duplication over dependency than shared error type here

// TODO: it is a lot more better to use non global way to manage error details

// TODO: DetailError should not be true by default

var EnableDetailError bool = true

const (
	SystemErrorKey ErrorCode = "generic.generic.SYSTEM_ERROR"

	SystemErrorText string = "system error occurred, contact admin."
)

type SystemError struct {
	id string

	err error
}

func NewSystemError(err error) SystemError {
	se := SystemError{err: err}

	se.id = uuid.NewString()

	return se
}

func NewManagedSystemError(err error, id string) SystemError {
	return SystemError{err: err, id: id}
}

func (se SystemError) Error() string {
	dt := ErrorDto{
		Id: se.id,

		Description: SystemErrorText,

		ErrorCode: SystemErrorKey,

		Type: ErrorTypeSystemError,
	}

	if EnableDetailError {
		dt.Description = se.err.Error()
	}

	raw, err := json.Marshal(dt)
	if err != nil {
		panic(err)
	}

	return string(raw)
}
