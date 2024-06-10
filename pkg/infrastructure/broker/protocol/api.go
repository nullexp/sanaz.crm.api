package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	protError "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/log"
)

type ReturnMessage[T any] struct {
	ErrorMessage string `json:"errorMessage"`
	Response     T      `json:"response"`
}

type RequestMessage[RequestType any] struct {
	Request RequestType `json:"request"`
}

func (rm RequestMessage[RequestType]) GetRequest() RequestType {
	return rm.Request
}

func NewReturnMessage[T any](message T) ReturnMessage[T] {
	return ReturnMessage[T]{Response: message}
}

func NewErrorReturnMessage[T any](err error) ReturnMessage[T] {
	return ReturnMessage[T]{ErrorMessage: err.Error()}
}

func answer(ctx context.Context, msg ResponseMessage, data []byte) {
	answer := msg.Answer(ctx, data)
	err := <-answer
	if err != nil {
		log.LoggerInstance.Error(err)
	}
}

// TODO: the binary data is fixed and can be cached
func GetProtocolNegotiationFailedData() []byte {
	nm := NewErrorReturnMessage[string](ErrProtocolNegotiationFailed)
	marshalData, _ := json.Marshal(nm) // Will never fail
	return marshalData
}

// Handler and requester with single param and single response

type BasicHandler[ReturnType, Param any] func(ctx context.Context, params Param) (out ReturnType, err error)

func RespondBasicHandler[ReturnType, Param any](responder SubscribeToResponder, topic string, handler BasicHandler[ReturnType, Param]) error {
	dataChannel, err := responder.SubscribeToRespond(context.Background(), topic)
	if err != nil {
		return err
	}

	go func(dataChannel <-chan ResponseMessage, topic string) {
		for msg := range dataChannel {
			// TODO: there must be a timeout param for create a context with timeout
			go respondBasicHandler[ReturnType, Param](context.Background(), msg, handler, topic)
		}
	}(dataChannel, topic)

	return nil
}

func respondBasicHandler[ReturnType, Param any](ctx context.Context, msg ResponseMessage, handler BasicHandler[ReturnType, Param], topic string) {
	data := msg.GetData()
	requestMessage := new(RequestMessage[Param])
	err := json.Unmarshal(data, requestMessage)
	if err != nil {
		answer(ctx, msg, GetProtocolNegotiationFailedData())
		return
	}

	params := requestMessage.GetRequest()
	var out ReturnMessage[ReturnType]
	result, err := handler(ctx, params)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, params).Errorf("Failed request: %v", err)
		out = NewErrorReturnMessage[ReturnType](err)
	} else {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, params).Infof("Succesed request: %v", result)
		out = NewReturnMessage[ReturnType](result)
	}

	marshalData, err := json.Marshal(out)
	if err != nil {
		return
	}

	answer(ctx, msg, marshalData)
}

func RequestBasicHandler[ReturnType, Param any](requester Requester, topic string, param Param) (out ReturnType, err error) {
	dto := RequestMessage[Param]{Request: param}
	raw, err := json.Marshal(dto)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}

	ctx := context.Background()
	data := requester.Request(ctx, topic, raw)
	dt := <-data
	if dt.Error != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		// TODO: use better error id
		err = protError.NewManagedSystemError(dt.Error, uuid.NewString())
		return
	}

	returnMessage := new(ReturnMessage[ReturnType])
	err = json.Unmarshal(dt.Value, returnMessage)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}

	if returnMessage.ErrorMessage != "" {
		err = errors.New(returnMessage.ErrorMessage)
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}

	log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Infof("Succeed request: %v", returnMessage.Response)
	out = returnMessage.Response
	return
}

// Handler with single param and no return type

type DeleteHandler[Param any] func(ctx context.Context, params Param) (err error)

type UpdateHandler[Param any] func(ctx context.Context, params Param) (err error)

func RespondDeleteHandler[Param any](responder SubscribeToResponder, topic string, handler DeleteHandler[Param]) error {
	dataChannel, err := responder.SubscribeToRespond(context.Background(), topic)
	if err != nil {
		return err
	}

	go func(dataChannel <-chan ResponseMessage, topic string) {
		for msg := range dataChannel {
			// TODO: there must be a timeout param for create a context with timeout
			go respondDeleteHandler[Param](context.Background(), msg, handler, topic)
		}
	}(dataChannel, topic)

	return nil
}

func RespondUpdateHandler[Param any](responder SubscribeToResponder, topic string, handler UpdateHandler[Param]) error {
	return RespondDeleteHandler[Param](responder, topic, DeleteHandler[Param](handler))
}

func respondDeleteHandler[Param any](ctx context.Context, msg ResponseMessage, handler DeleteHandler[Param], topic string) {
	data := msg.GetData()
	requestMessage := new(RequestMessage[Param])
	err := json.Unmarshal(data, requestMessage)
	if err != nil {
		answer(ctx, msg, GetProtocolNegotiationFailedData())
		return
	}

	var out ReturnMessage[bool]
	params := requestMessage.GetRequest()
	// Since we want to use one type of return message, we set return type to boolean
	err = handler(ctx, params)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, params).Errorf("Failed request: %v", err)
		out = NewErrorReturnMessage[bool](err)
	} else {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, params).Infof("Succeed request: %v", true)
		out = NewReturnMessage[bool](true)
	}

	marshalData, err := json.Marshal(out)
	if err != nil {
		return
	}

	answer(ctx, msg, marshalData)
}

func RequestDeleteHandler[Param any](requester Requester, topic string, param Param) (err error) {
	dto := RequestMessage[Param]{Request: param}
	raw, err := json.Marshal(dto)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}
	ctx := context.Background()
	data := requester.Request(ctx, topic, raw)
	dt := <-data
	if dt.Error != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		// TODO: use better error id
		err = protError.NewManagedSystemError(dt.Error, uuid.NewString())
		return
	}

	returnMessage := new(ReturnMessage[bool])
	err = json.Unmarshal(dt.Value, returnMessage)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}
	if returnMessage.ErrorMessage != "" {
		err = errors.New(returnMessage.ErrorMessage)
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Errorf("Failed request: %v", err)
		return
	}

	log.LoggerInstance.WithTime(time.Now()).WithField(topic, param).Info("Succeed request")
	return
}

func RequestUpdateHandler[Param any](requester Requester, topic string, param Param) (err error) {
	return RequestDeleteHandler[Param](requester, topic, param)
}

// Handler with no param and a return type

type GetHandler[ReturnType any] func(ctx context.Context) (out ReturnType, err error)

func RespondGetHandler[ReturnType any](responder SubscribeToResponder, topic string, handler GetHandler[ReturnType]) error {
	dataChannel, err := responder.SubscribeToRespond(context.Background(), topic)
	if err != nil {
		return err
	}

	go func(dataChannel <-chan ResponseMessage, topic string) {
		for msg := range dataChannel {
			// TODO: there must be a timeout param for create a context with timeout
			go respondGetHandler[ReturnType](context.Background(), msg, handler, topic)
		}
	}(dataChannel, topic)

	return nil
}

func respondGetHandler[ReturnType any](ctx context.Context, msg ResponseMessage, handler GetHandler[ReturnType], topic string) {
	data := msg.GetData()
	requestMessage := new(RequestMessage[bool])
	err := json.Unmarshal(data, requestMessage)
	if err != nil {
		answer(ctx, msg, GetProtocolNegotiationFailedData())
		return
	}

	var out ReturnMessage[ReturnType]
	result, err := handler(ctx)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Errorf("Failed request: %v", err)
		out = NewErrorReturnMessage[ReturnType](err)
	} else {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Infof("Succesed request: %v", result)
		out = NewReturnMessage[ReturnType](result)
	}

	marshalData, err := json.Marshal(out)
	if err != nil {
		return
	}

	answer(ctx, msg, marshalData)
}

func RequestGetHandler[ReturnType any](requester Requester, topic string) (out ReturnType, err error) {
	// We use boolean as no param type, we don't want to have multiple request message for now
	dto := RequestMessage[bool]{}
	raw, err := json.Marshal(dto)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Errorf("Failed request: %v", err)
		return
	}

	ctx := context.Background()
	data := requester.Request(ctx, topic, raw)
	dt := <-data
	if dt.Error != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Errorf("Failed request: %v", err)
		// TODO: use better error id
		err = protError.NewManagedSystemError(dt.Error, uuid.NewString())
		return
	}

	returnMessage := new(ReturnMessage[ReturnType])
	err = json.Unmarshal(dt.Value, returnMessage)
	if err != nil {
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Errorf("Failed request: %v", err)
		return
	}

	if returnMessage.ErrorMessage != "" {
		err = errors.New(returnMessage.ErrorMessage)
		log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Errorf("Failed request: %v", err)
		return
	}

	log.LoggerInstance.WithTime(time.Now()).WithField(topic, nil).Infof("Succeed request: %v", returnMessage.Response)
	out = returnMessage.Response
	return
}

// Handler with no param and a return type

type NotifyHandler func(ctx context.Context) (err error)

func RespondNotifyHandler(responder SubscribeToResponder, topic string, handler NotifyHandler) error {
	dataChannel, err := responder.SubscribeToRespond(context.Background(), topic)
	if err != nil {
		return err
	}

	go func(dataChannel <-chan ResponseMessage) {
		for msg := range dataChannel {
			// TODO: there must be a timeout param for create a context with timeout
			go respondNotifyHandler(context.Background(), msg, handler)
		}
	}(dataChannel)

	return nil
}

func respondNotifyHandler(ctx context.Context, msg ResponseMessage, handler NotifyHandler) {
	data := msg.GetData()
	requestMessage := new(RequestMessage[bool])
	err := json.Unmarshal(data, requestMessage)
	if err != nil {
		answer(ctx, msg, GetProtocolNegotiationFailedData())
		return
	}
	var out ReturnMessage[bool]
	err = handler(ctx)
	if err != nil {
		out = NewErrorReturnMessage[bool](err)
	} else {
		out = NewReturnMessage[bool](true)
	}

	marshalData, err := json.Marshal(out)
	if err != nil {
		return
	}

	answer(ctx, msg, marshalData)
}

func RequestNotifyHandler(requester Requester, topic string) (err error) {
	// We use boolean as no param type, we don't want to have multiple request message for now
	dto := RequestMessage[bool]{}
	raw, err := json.Marshal(dto)
	if err != nil {
		return
	}
	ctx := context.Background()
	data := requester.Request(ctx, topic, raw)
	dt := <-data
	if dt.Error != nil {
		// TODO: use better error id
		err = protError.NewManagedSystemError(dt.Error, uuid.NewString())
		return
	}

	returnMessage := new(ReturnMessage[bool])
	err = json.Unmarshal(dt.Value, returnMessage)
	if err != nil {
		return
	}
	if returnMessage.ErrorMessage != "" {
		err = errors.New(returnMessage.ErrorMessage)
		return
	}
	return
}
