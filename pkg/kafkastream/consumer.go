package kafkastream

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/ashep/joex/pkg/eventstream"
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ConsumerConfig struct {
	brokers              []string
	topics               []string
	group                string
	autoCommit           bool
	blockRebalanceOnPoll bool
	maxAttempts          int
	l                    zerolog.Logger
}

type ConsumerOption func(cfg *ConsumerConfig)

func WithConsumerBrokers(brokers []string) ConsumerOption {
	return func(cfg *ConsumerConfig) {
		cfg.brokers = brokers
	}
}

func WithConsumerTopics(topics []string) ConsumerOption {
	return func(cfg *ConsumerConfig) {
		cfg.topics = slices.Clone(topics)
	}
}

func WithConsumerGroup(group string) ConsumerOption {
	return func(cfg *ConsumerConfig) {
		cfg.group = group
	}
}

func WithConsumerLogger(l zerolog.Logger) ConsumerOption {
	return func(cfg *ConsumerConfig) {
		cfg.l = l
	}
}

func WithConsumerAutoCommit() ConsumerOption {
	return func(cfg *ConsumerConfig) {
		cfg.autoCommit = true
	}
}

type Consumer[T any] struct {
	cli         *kgo.Client
	maxAttempts int
	attempts    map[string]int
	l           zerolog.Logger
}

func NewConsumer[T any](opts ...ConsumerOption) (*Consumer[T], error) {
	cfg := &ConsumerConfig{
		brokers:              []string{"127.0.0.1:9092"},
		topics:               []string{"default"},
		group:                "default",
		autoCommit:           false,
		blockRebalanceOnPoll: true,
		maxAttempts:          3,
		l:                    zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(cfg.brokers...),
		kgo.ConsumeTopics(cfg.topics...),
		kgo.ConsumerGroup(cfg.group),
	}
	if !cfg.autoCommit {
		kgoOpts = append(kgoOpts, kgo.DisableAutoCommit())
	}
	if cfg.blockRebalanceOnPoll {
		kgoOpts = append(kgoOpts, kgo.BlockRebalanceOnPoll())
	}

	cli, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Consumer[T]{
		cli:         cli,
		maxAttempts: cfg.maxAttempts,
		attempts:    make(map[string]int),
		l:           cfg.l,
	}, nil
}

func (c *Consumer[T]) Consume(ctx context.Context, proc func(context.Context, *eventstream.ConsumedEvent[T]) error) {
	var sleep time.Duration

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}

		fetches := c.cli.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				c.l.Err(err.Err).Msg("poll fetches failed")
			}
			sleep = time.Second
			continue
		}

		sleep = 0
		rewindOffsets := make(map[string]map[int32]kgo.EpochOffset)

		iter := fetches.RecordIter()
		for !iter.Done() {
			rec := iter.Next()
			c.l.Debug().
				Str("topic", rec.Topic).
				Int("partition", int(rec.Partition)).
				Str("key", string(rec.Key)).
				Msg("record fetched")

			v := new(T)
			if err := json.Unmarshal(rec.Value, v); err != nil {
				c.handleProcessError(rec, fmt.Errorf("unmarshal: %w", err), rewindOffsets)
			}

			pErr := proc(ctx, &eventstream.ConsumedEvent[T]{
				Topic:     rec.Topic,
				Partition: int(rec.Partition),
				Offset:    int(rec.Offset),
				Key:       rec.Key,
				RawValue:  rec.Value,
				Value:     *v,
			})

			if pErr != nil {
				c.handleProcessError(rec, pErr, rewindOffsets)
			}
		}

		if len(rewindOffsets) > 0 {
			c.l.Warn().Interface("offsets", rewindOffsets).Msg("rewind offsets")
			c.cli.SetOffsets(rewindOffsets)
		}

		if err := c.cli.CommitUncommittedOffsets(ctx); err != nil {
			c.l.Err(err).Msg("commit uncommitted offsets failed")
		} else {
			c.l.Debug().Msg("uncommitted offsets committed")
		}

		c.cli.AllowRebalance()
	}
}

func (c *Consumer[T]) handleProcessError(rec *kgo.Record, err error, rewindOffsets map[string]map[int32]kgo.EpochOffset) {
	c.l.Warn().
		Err(err).
		Str("key", string(rec.Key)).
		Int("partition", int(rec.Partition)).
		Int("offset", int(rec.Offset)).
		Int("attempt", c.attempts[string(rec.Key)]).
		Msg("failed to process record")

	c.attempts[string(rec.Key)]++
	if c.attempts[string(rec.Key)] == c.maxAttempts {
		delete(c.attempts, string(rec.Key))
		c.l.Error().Str("key", string(rec.Key)).Msg("max processing attempts reached, skipping record")
		return
	}

	if rewindOffsets[rec.Topic] == nil {
		rewindOffsets[rec.Topic] = make(map[int32]kgo.EpochOffset)
	}
	rewindOffsets[rec.Topic][rec.Partition] = kgo.EpochOffset{
		Epoch:  rec.LeaderEpoch,
		Offset: rec.Offset,
	}
}

func (c *Consumer[T]) Close() {
	c.cli.Close()
}
