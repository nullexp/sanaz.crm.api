package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	const detail = "detail"

	const description = "describing current error"

	const codeName = "TestService.User.USER_ALREADY_EXIST"

	const id = "1f38b18b-2606-49dc-99b0-ed187e0a2618"

	testError := NewNotFoundError(ErrorCode(codeName), description)

	modified := testError.WithIdAndDetail(id, detail)

	assert.EqualValues(t, len(testError.Details), 0)

	assert.EqualValues(t, len(modified.Details), 1)

	parsedError := "{\"id\":\"1f38b18b-2606-49dc-99b0-ed187e0a2618\",\"errorCode\":\"TestService.User.USER_ALREADY_EXIST\",\"description\":\"describing current error\",\"details\":[\"detail\"],\"type\":\"NotFound\"}"

	assert.Equal(t, parsedError, modified.Error())
}
