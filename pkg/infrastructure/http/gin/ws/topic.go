package ws

// reserved topics

const (
	ConnectionWasForTooLong    = "connection deadline occured! server will disconnect the connection after given time"
	ConnectionWasForJwtExpired = "connection deadline occured! server will disconnect the connection after jwt expired"
)

// Error codes
const (
	TimeoutError = "timeout-error"
	TokenError   = "jwt-expired"
)

// Topics
const (
	ErrorTopic = "error"
)

func GetReservedTopic() []string {
	return []string{ErrorTopic}
}
