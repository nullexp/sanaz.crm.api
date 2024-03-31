package sms

import "time"

type SMSNotifierKind string

const (
	SMSNotifierKindTest      SMSNotifierKind = "test"
	SMSNotifierKindKavenegar SMSNotifierKind = "kavenegar"
)

// TODO: add case test.
func NewSMSNotifier(kind SMSNotifierKind, params ...any) SMSNotifier {
	switch kind {
	case SMSNotifierKindKavenegar:
		return NewKavenegarSMSNotifier(params[0].(string), params[1].(time.Duration), params[2].(time.Duration))
	case SMSNotifierKindTest:
		panic("not inplemented")
	}
	panic("not implemented")
}
