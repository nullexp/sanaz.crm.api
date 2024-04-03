package misc

import "fmt"

type UserOperationError struct {
	ErrorCode string
	NotFound  bool
}

func (uo UserOperationError) GetOperationErrorCode() string {
	return fmt.Sprintf(uo.ErrorCode)
}

func (uo UserOperationError) IsNotFoundError() bool {
	return uo.NotFound
}

func (uo UserOperationError) Error() string {
	return fmt.Sprintf(OEOccured, uo.ErrorCode)
}

type OperationErrorGetter interface {
	GetOperationErrorCode() string
	IsNotFoundError() bool
}

const OEOccured = "Operation error occured with code %s"

func ToUserOperationError(e error) (ok bool, oe OperationErrorGetter) {
	oe, ok = e.(OperationErrorGetter)
	return
}

func WrapUserOperationError(e string) error {
	return UserOperationError{ErrorCode: e}
}

func WrapUserOperationNotFoundError(e string) error {
	return UserOperationError{ErrorCode: e, NotFound: true}
}
