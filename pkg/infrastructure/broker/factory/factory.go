package factory

import (
	mqtt "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/broker/mqtt"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/broker/protocol"
	redis "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/broker/redis"
)

type BrokerType string

const (
	MqttPaho = "Mqtt-Paho"

	Nats = "Nats"

	Redis = "Redis"

	Test = "Test"
)

// TODO: use type for option
// TODO: add nats
func NewBroker(name BrokerType, params ...any) protocol.Broker {
	switch name {

	case Test:
		panic("not implemented")

	case Redis:
		return redis.NewRedisClient(params[0].(string), params[1].(string), params[2].(string), params[3].(string))

	case MqttPaho:
		return mqtt.NewMqttClient(params[0].(string), params[1].(string),
			params[2].(string), params[3].(string), params[4].(byte), params[5].(bool))

	}

	panic("not implemented")
}

// TODO: use type for option

func NewFlusherBroker(name BrokerType, params ...any) protocol.FlusherBroker {
	switch name {

	case Test:
		panic("not implemented")

	case Redis:
		return redis.NewRedisClient(params[0].(string), params[1].(string), params[2].(string), params[3].(string))

	case MqttPaho:
		return mqtt.NewMqttClient(params[0].(string), params[1].(string), params[2].(string),
			params[3].(string), params[4].(byte), params[5].(bool))

	}

	panic("not implemented")
}
