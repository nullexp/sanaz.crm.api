package protocol

func IsManagedError(err error) (bool, OperationErrorGetter) {
	if uo, ok := err.(UserOperationError); ok {
		return true, OperationErrorGetter{OperationErrorCode: (uo.ErrorCode), IsNotFound: false}
	}
	if nf, ok := err.(NotFoundError); ok {
		return true, OperationErrorGetter{OperationErrorCode: (nf.ErrorCode), IsNotFound: true}
	}
	if _, ok := err.(SystemError); ok {
		return true, OperationErrorGetter{OperationErrorCode: SystemErrorKey, IsNotFound: false}
	}
	return false, OperationErrorGetter{}
}

type OperationErrorGetter struct {
	OperationErrorCode ErrorCode
	IsNotFound         bool
}

const OEOccured = "Operation error occured with code %s"

func WrapUserOperationError(code ErrorCode, description string) error {
	return UserOperationError{ErrorCode: code, Description: description}
}

func WrapUserOperationNotFoundError(e ErrorCode, description string) error {
	return NotFoundError{ErrorCode: e, Description: description}
}
