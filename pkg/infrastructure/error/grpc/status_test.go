package grpc

import (
	"testing"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
	"github.com/stretchr/testify/assert"
)

func TestGrpcStatus(t *testing.T) {
	const description = "describing current error"

	const codeName = "TestService.User.USER_ALREADY_EXIST"

	testError := protocol.NewUserOperationError(codeName, description)

	er := ToGrpcError(testError)

	assert.EqualValues(t, "rpc error: code = InvalidArgument desc = {\"id\":\"\",\"errorCode\":\"TestService.User.USER_ALREADY_EXIST\",\"description\":\"describing current error\",\"type\":\"UserOperation\"}", er.Error())
}
