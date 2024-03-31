package protocol

import "context"

type (
	Publisher interface {
		Publish(ctx context.Context, subject string, content []byte) (err <-chan error)
	}

	Response struct {
		Error error
		Value []byte
	}
	Requester interface {
		Request(ctx context.Context, subject string, content []byte) (value <-chan Response)
	}

	ResponseMessage interface {
		GetData() []byte
		Answer(ctx context.Context, content []byte) (err <-chan error)
	}

	SubscribeToResponder interface {
		SubscribeToRespond(ctx context.Context, subject string) (value <-chan ResponseMessage, err error)
	}

	Subscriber interface {
		Subscribe(ctx context.Context, subject string) (value <-chan []byte, err error)
	}

	Unsubscriber interface {
		Unsubscribe(ctx context.Context, subject string) (err error)
	}

	Connecter interface {
		Connect() error
	}

	Disconnecter interface {
		Disconnect() error
	}

	Flusher interface {
		Flush(ctx context.Context) error
	}

	Broker interface {
		Publisher
		Subscriber
		Unsubscriber
		Connecter
		Disconnecter
	}
	RequesterBroker interface {
		Broker
		Requester
		SubscribeToResponder
	}
	FlusherBroker interface {
		Broker
		Flusher
	}
	FlusherRequester interface {
		RequesterBroker
		Flusher
	}
)
