package protocol

import "errors"

var (
	ErrTimeout                   = errors.New("timeout")
	ErrProtocolNegotiationFailed = errors.New("protocol negotiation failed, expected message cannot be marshaled")
)
