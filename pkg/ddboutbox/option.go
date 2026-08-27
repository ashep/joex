package ddboutbox

import (
	"time"

	"github.com/rs/zerolog"
)

type config struct {
	pollPeriod time.Duration
	l          zerolog.Logger
}

type Option func(*config)

func WithPollPeriod(p time.Duration) Option {
	if p <= time.Second {
		p = time.Second
	}

	return func(o *config) {
		o.pollPeriod = p
	}
}

func WithLogger(l zerolog.Logger) Option {
	return func(o *config) {
		o.l = l
	}
}
