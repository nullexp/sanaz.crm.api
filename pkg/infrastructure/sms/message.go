package sms

type Message struct {
	PhoneNumber, Text string
}

type SMSNotifier interface {
	SendMessage(message Message) error
	SendBulkMessages(messages []Message) error             // will return error if one fail
	SendBulkMessagesAsync(messages []Message) <-chan error // will return error if one fail
}
