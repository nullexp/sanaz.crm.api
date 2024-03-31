package protocol

func IsManagedError(err error) bool {
	if _, ok := err.(UserOperationError); ok {
		return true
	}
	if _, ok := err.(NotFoundError); ok {
		return true
	}
	if _, ok := err.(SystemError); ok {
		return true
	}
	return false
}
