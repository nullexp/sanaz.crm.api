package nats

import (
	"context"
	"errors"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/broker/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/log"
	"github.com/nats-io/nats.go"
)

func init() {
	log.LoggerInstance = log.NewLog(log.LogLevel)
}

type subpub struct {
	NatsPubSub   *nats.Subscription
	reponderChan chan protocol.ResponseMessage
	rawChannel   chan []byte
}

type NatsClient struct {
	url      string
	client   *nats.Conn
	subsChan map[string]*subpub
	timeout  time.Duration
}

type natsResponder struct {
	message []byte
	respond func(content []byte) error
}

func NewNatsWithClientTimeout(client *nats.Conn, timeout time.Duration) protocol.FlusherRequester {
	return &NatsClient{client: client, subsChan: make(map[string]*subpub), timeout: timeout}
}

func NewNatsWithClient(client *nats.Conn) protocol.FlusherRequester {
	return &NatsClient{client: client, subsChan: make(map[string]*subpub), timeout: 5 * time.Second}
}

func NewNatsClient(url string) protocol.FlusherRequester {
	return &NatsClient{url: url, subsChan: make(map[string]*subpub), timeout: 5 * time.Second}
}

func (n natsResponder) GetData() []byte {
	return n.message
}

func (n natsResponder) Answer(ctx context.Context, content []byte) <-chan error {
	errValue := make(chan error)
	go func() {
		err := n.respond(content)
		errValue <- err
	}()

	return errValue
}

func (n *NatsClient) SubscribeToRespond(ctx context.Context, subject string) (value <-chan protocol.ResponseMessage, err error) {
	brokerChan := make(chan protocol.ResponseMessage)
	// Subscribe to the subject using the channel
	sub, err := n.client.Subscribe(subject, func(m *nats.Msg) {
		brokerChan <- natsResponder{message: m.Data, respond: m.Respond}
	})
	if err != nil {
		return nil, err
	}

	n.subsChan[subject] = &subpub{
		NatsPubSub:   sub,
		reponderChan: brokerChan,
	}

	return brokerChan, nil
}

func (n *NatsClient) Request(ctx context.Context, subject string, content []byte) (value <-chan protocol.Response) {
	response := make(chan protocol.Response)
	// ignoring ctx
	go func() {
		rs := protocol.Response{}
		msg, err := n.client.Request(subject, content, n.timeout)
		if err != nil {
			rs.Error = err
			response <- rs
		} else {
			rs.Value = msg.Data
			response <- rs
		}
	}()

	return response
}

func (n *NatsClient) Publish(ctx context.Context, subject string, content []byte) (err <-chan error) {
	signal := make(chan error)
	go func() {
		err := n.client.Publish(subject, content)
		if err != nil {
			signal <- err
		} else {
			signal <- nil
		}
	}()

	return signal
}

func (n *NatsClient) Subscribe(ctx context.Context, subject string) (<-chan []byte, error) {
	// Use a channel to receive messages
	messageCh := make(chan []byte)
	// Subscribe to the subject using the channel
	sub, err := n.client.Subscribe(subject, func(m *nats.Msg) {
		messageCh <- m.Data
	})

	n.subsChan[subject] = &subpub{
		NatsPubSub: sub,
		rawChannel: messageCh,
	}

	return messageCh, err
}

func (n *NatsClient) Unsubscribe(ctx context.Context, subject string) (err error) {
	if v, ok := n.subsChan[subject]; ok {
		err := v.NatsPubSub.Unsubscribe()
		delete(n.subsChan, subject)
		if err != nil {
			return err
		}
	}

	return errors.New("subject no exist")
}

func (n *NatsClient) Connect() error {
	nc, err := nats.Connect(n.url)
	if err != nil {
		return err
	}
	n.client = nc

	return nil
}

func (n *NatsClient) Disconnect() error {
	for k, v := range n.subsChan {
		if v.reponderChan != nil {
			close(v.reponderChan)
		}
		if v.rawChannel != nil {
			close(v.reponderChan)
		}
		delete(n.subsChan, k)
	}

	if n.client == nil {
		return errors.New("nats connection is nil")
	}

	return n.client.Drain()
}

func (n *NatsClient) Flush(ctx context.Context) error {
	return n.client.Flush()
}
