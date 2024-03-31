package error

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
	"github.com/go-playground/validator"
)

const (
	GenericValidationErrorKey = "generic.generic.VALIDATION_ERROR"

	GenericValidationErrorDesc = "user input is invalid, please check values"
)

var ErrGenericValidationError = protocol.NewUserOperationError(GenericValidationErrorKey, GenericValidationErrorDesc)

func Validate(ctx context.Context, dto any) error {
	validate := validator.New()

	err := validate.StructCtx(ctx, dto)

	if err == nil {
		return nil
	}

	return ErrGenericValidationError.WithDetail(err.Error())
}

func IsManagedError(err error) bool {
	if _, ok := err.(protocol.UserOperationError); ok {
		return true
	} else if _, ok := err.(protocol.NotFoundError); ok {
		return true
	} else if _, ok := err.(protocol.SystemError); !ok {
		return true
	}

	return false
}
