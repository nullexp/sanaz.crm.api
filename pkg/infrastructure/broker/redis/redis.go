package redis

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/broker/protocol"
	redis "github.com/redis/go-redis/v9"
)

type subpub struct {
	redisPubSub *redis.PubSub

	channel chan []byte
}

type RedisClient struct {
	username, password, fullAddress, clientName string

	client *redis.Client

	subsChan map[string]*subpub
}

func NewRedisWithClient(client *redis.Client) protocol.FlusherBroker {
	return &RedisClient{client: client, subsChan: make(map[string]*subpub)}
}

func NewRedisClient(username, password, clientName, fullAddress string) protocol.FlusherBroker {
	return &RedisClient{username: username, password: password, fullAddress: fullAddress, clientName: clientName, subsChan: make(map[string]*subpub)}
}

func (rc *RedisClient) Publish(ctx context.Context, subject string, content []byte) (err <-chan error) {
	signal := make(chan error)

	go func() {
		err := rc.client.Publish(ctx, subject, content).Err()

		signal <- err
	}()

	return signal
}

func (rc *RedisClient) Subscribe(ctx context.Context, subject string) (value <-chan []byte, err error) {
	// TODO: Client does not really support subscribe assurence or I can't find it

	ping := rc.client.Ping(ctx)

	err = ping.Err()

	if ping.Err() != nil {
		return
	}

	outValue := make(chan []byte)

	pubsub := rc.client.Subscribe(ctx, subject)

	go func() {
		for v := range pubsub.Channel() {
			outValue <- []byte(v.Payload)
		}
	}()

	// TODO: not thread safe

	rc.subsChan[subject] = &subpub{redisPubSub: pubsub, channel: outValue}

	return outValue, nil
}

func (rc *RedisClient) Unsubscribe(ctx context.Context, subject string) (err error) {
	chn := rc.subsChan[subject]

	if chn == nil {
		return
	}

	err = chn.redisPubSub.Unsubscribe(ctx, subject)

	close(chn.channel)

	delete(rc.subsChan, subject)

	return
}

func (rc *RedisClient) Connect() error {
	ping := func() error {
		return rc.client.Ping(context.Background()).Err()
	}

	if rc.client != nil {
		return ping()
	}

	rc.client = redis.NewClient(&redis.Options{
		Addr: rc.fullAddress,

		Password: rc.password,

		Username: rc.username,

		ClientName: rc.clientName,
	})

	return ping()
}

func (rc *RedisClient) Disconnect() error {
	for k, v := range rc.subsChan {

		close(v.channel)

		delete(rc.subsChan, k)

	}

	return rc.client.Close()
}

func (rc *RedisClient) Flush(ctx context.Context) error {
	// Redis PUBSUB does not require flush

	return nil
}
