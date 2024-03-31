package grpc

import (
	"encoding/json"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/error/protocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGrpcError(err error) error {
	var code codes.Code = codes.Internal

	if _, ok := err.(protocol.UserOperationError); ok {
		code = codes.InvalidArgument
	} else if _, ok := err.(protocol.NotFoundError); ok {
		code = codes.NotFound
	} else if _, ok := err.(protocol.SystemError); !ok {
		err = protocol.NewSystemError(err)
	}

	return status.Error(code, err.Error())
}

func ToBffError(err error) error {
	if _, ok := err.(protocol.UserOperationError); ok {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if _, ok := err.(protocol.NotFoundError); ok {
		return status.Error(codes.NotFound, err.Error())
	} else if _, ok := err.(protocol.SystemError); ok {
		return status.Error(codes.Internal, err.Error())
	}

	strValue := err.Error()

	dto := protocol.ErrorDto{}

	marshalError := json.Unmarshal([]byte(strValue), &dto)

	if marshalError != nil {

		err = protocol.NewSystemError(err)

		return status.Error(codes.Internal, err.Error())

	}

	switch dto.Type {

	case protocol.ErrorTypeUserOperation:

		return status.Error(codes.InvalidArgument, err.Error())

	case protocol.ErrorTypeNotFound:

		return status.Error(codes.NotFound, err.Error())

	}

	return status.Error(codes.Internal, err.Error())
}
