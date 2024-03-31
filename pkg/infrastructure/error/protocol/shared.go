package protocol

type (
	ErrorCode string
	ErrorType string
)

const (
	ErrorTypeUserOperation ErrorType = "UserOperation"
	ErrorTypeNotFound      ErrorType = "NotFound"
	ErrorTypeSystemError   ErrorType = "SystemError"
)

type ErrorDto struct {
	Id          string        `json:"id"`
	ErrorCode   ErrorCode     `json:"errorCode"`
	Description string        `json:"description"`
	Details     []interface{} `json:"details,omitempty"`
	Type        ErrorType     `json:"type"`
}
