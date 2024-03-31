package protocol

import "context"

type (
	Connecter interface {
		Connect() error
	}

	Disconnecter interface {
		Disconnect() error
	}

	RawSetter interface {
		Set(ctx context.Context, key string, value []byte) error
	}

	RawFetcher interface {
		Fetch(ctx context.Context, key string) ([]byte, error)
	}

	Deleter interface {
		Delete(ctx context.Context, key string) error
	}

	RawCacheStore interface {
		RawSetter
		RawFetcher
		Deleter
		Connecter
		Disconnecter
	}
)
