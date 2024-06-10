//nolint:all
package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/broker/protocol"
)

type memoryResponder struct {
	message []byte

	respond func(content []byte) error
}

type MemoryClient struct {
	subsLock     sync.RWMutex
	subscribers  map[string][]chan []byte
	requestsLock sync.RWMutex
	requests     map[string]chan []byte
}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		subscribers: make(map[string][]chan []byte),
		requests:    make(map[string]chan []byte),
	}
}

func (m memoryResponder) GetData() []byte {
	return m.message
}

func (m memoryResponder) Answer(ctx context.Context, content []byte) <-chan error {
	errValue := make(chan error)

	go func() {
		err := m.respond(content)

		errValue <- err
	}()

	return errValue
}

func (mc *MemoryClient) Publish(ctx context.Context, subject string, content []byte) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		mc.subsLock.RLock()
		defer mc.subsLock.RUnlock()

		channels, ok := mc.subscribers[subject]
		if !ok {
			errCh <- fmt.Errorf("there is no subscriber for this subject: %s", subject)
			return
		}

		for _, ch := range channels {
			go func(c chan []byte) {
				c <- content
				fmt.Println("published on subject: " + subject)
			}(ch)
		}

		errCh <- nil
	}()

	return errCh
}

func (mc *MemoryClient) Subscribe(ctx context.Context, subject string) (<-chan []byte, error) {
	subCh := make(chan []byte)

	mc.subsLock.Lock()
	mc.subscribers[subject] = append(mc.subscribers[subject], subCh)
	mc.subsLock.Unlock()

	return subCh, nil
}

func (mc *MemoryClient) Unsubscribe(ctx context.Context, subject string) error {
	mc.subsLock.Lock()
	defer mc.subsLock.Unlock()

	channels, ok := mc.subscribers[subject]
	if !ok {
		return errors.New("subject not exist : " + subject)
	}

	for _, c := range channels {
		close(c)
	}

	return nil
}

func (mc *MemoryClient) Request(ctx context.Context, subject string, content []byte) <-chan protocol.Response {
	responseCh := make(chan protocol.Response, 1)

	mc.requestsLock.Lock()
	mc.requests[subject] = make(chan []byte)
	mc.requestsLock.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			mc.requestsLock.Lock()
			delete(mc.requests, subject)
			mc.requestsLock.Unlock()
			responseCh <- protocol.Response{Error: ctx.Err()}
		case respond := <-mc.requests[subject]:
			responseCh <- protocol.Response{Value: respond}
		}
	}()

	go func() {
		errCh := mc.Publish(ctx, subject, content)
		select {
		case err := <-errCh:
			if err != nil {
				responseCh <- protocol.Response{Error: err}
			}
		}
	}()

	return responseCh
}

func (mc *MemoryClient) SubscribeToRespond(ctx context.Context, subject string) (<-chan protocol.ResponseMessage, error) {
	respCh := make(chan protocol.ResponseMessage)

	subCh, err := mc.Subscribe(ctx, subject)
	if err != nil {
		return nil, err
	}

	go func() {
		for msg := range subCh {
			select {
			case <-ctx.Done():
				return
			case respCh <- memoryResponder{
				message: msg,
				respond: mc.Respond,
			}:
			}
		}
	}()

	return respCh, nil
}

func (mc *MemoryClient) Respond(data []byte) error {
	mc.requestsLock.RLock()
	defer mc.requestsLock.RUnlock()

	for _, ch := range mc.requests {
		go func(c chan []byte) {
			c <- data
		}(ch)
	}

	return nil
}

func (mc *MemoryClient) Flush(ctx context.Context) error {
	return nil
}

func (mc *MemoryClient) Connect() error {
	return nil
}

func (mc *MemoryClient) Disconnect() error {
	return nil
}
