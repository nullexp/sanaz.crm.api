package Mqtt

import (
	"context"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/broker/protocol"
)

const defaultTimeOut = time.Second * 5

type MqttClient struct {
	option *mqtt.ClientOptions

	client mqtt.Client

	Qos byte
}

func NewMqttWithClient(client *mqtt.Client, qos byte) protocol.FlusherBroker {
	return &MqttClient{client: *client, Qos: qos}
}

func NewMqttClient(username, password, clientName, fullAddress string, qos byte, cleanSession bool) protocol.FlusherBroker {
	option := mqtt.NewClientOptions()

	option.AddBroker(fullAddress)

	option.SetClientID(clientName)

	// TODO: Default config should also be injected

	option.ConnectTimeout = time.Second * 5

	option.WriteTimeout = time.Second * 5

	option.KeepAlive = 10

	option.PingTimeout = time.Second * 5

	option.Username = username

	option.Password = password

	option.ConnectRetry = true

	option.AutoReconnect = true

	option.CleanSession = cleanSession

	option.WillQos = qos

	return &MqttClient{option: option, Qos: qos}
}

func (mc *MqttClient) Publish(ctx context.Context, subject string, content []byte) (err <-chan error) {
	signal := make(chan error)

	go func() {
		token := mc.client.Publish(subject, 1, false, content)

		send := token.WaitTimeout(defaultTimeOut)

		if !send {

			signal <- protocol.ErrTimeout

			return

		}

		signal <- token.Error()
	}()

	return signal
}

func (rc *MqttClient) Subscribe(ctx context.Context, subject string) (value <-chan []byte, err error) {
	outValue := make(chan []byte)

	token := rc.client.Subscribe(subject, 1, func(c mqtt.Client, m mqtt.Message) {
		outValue <- m.Payload()
	})

	return outValue, token.Error()
}

func (rc *MqttClient) Unsubscribe(ctx context.Context, subject string) (err error) {
	token := rc.client.Unsubscribe(subject)

	rs := token.WaitTimeout(defaultTimeOut)

	if !rs {
		return protocol.ErrTimeout
	}

	return token.Error()
}

func (rc *MqttClient) Connect() error {
	if rc.client == nil {
		rc.client = mqtt.NewClient(rc.option)
	}

	token := rc.client.Connect()

	// Excluding auto reconnect on fist connection

	timer := time.NewTimer(time.Second * 5)

	defer timer.Stop()

	select {

	case <-timer.C:

		return protocol.ErrTimeout

	case <-token.Done():

		return token.Error()

	}
}

const twoSecond = 2 * 1000

func (rc *MqttClient) Disconnect() error {
	if rc.client.IsConnectionOpen() {
		rc.client.Disconnect(twoSecond)
	}

	return nil
}

func (rc *MqttClient) Flush(ctx context.Context) error {
	// Mqtt  does not require flush

	return nil
}
