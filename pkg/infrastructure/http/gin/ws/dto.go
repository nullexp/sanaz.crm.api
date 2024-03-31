package ws

import (
	"encoding/json"

	"gitlab.espadev.ir/espad-go/infrastructure/misc"
)

type JsonDtoMessage struct {
	Topic string `json:"topic"`
	Dto   string `json:"dto"`
}

func NewJsonDtoMessage(topic string, dto any) (out JsonDtoMessage) {
	out.Topic = topic

	data, err := json.Marshal(dto)
	if err != nil {
		panic(err)
	}
	out.Dto = string(data)
	return out
}

type duplexMessage struct {
	Topic  string
	Dto    any
	Caller misc.Caller
}

func NewDuplexMessage(topic string, dto any, caller misc.Caller) duplexMessage {
	d := duplexMessage{Caller: caller}
	d.Topic = topic
	d.Dto = dto
	return d
}

func (w duplexMessage) GetDto() any {
	return w.Dto
}

func (w duplexMessage) GetTopic() string {
	return w.Topic
}

func (w duplexMessage) MustGetCaller() misc.Caller {
	return w.Caller
}
