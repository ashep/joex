package kafkastream

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ashep/joex/pkg/eventstream"
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ProducerConfig struct {
	brokers []string
	topic   string
	l       zerolog.Logger
}

type ProducerOption func(cfg *ProducerConfig)

func WithProducerBrokers(brokers []string) ProducerOption {
	return func(cfg *ProducerConfig) {
		cfg.brokers = slices.Clone(brokers)
	}
}

func WithProducerTopic(topic string) ProducerOption {
	return func(cfg *ProducerConfig) {
		cfg.topic = topic
	}
}

func WithProducerLogger(l zerolog.Logger) ProducerOption {
	return func(cfg *ProducerConfig) {
		cfg.l = l
	}
}

type Producer struct {
	cli   *kgo.Client
	topic string
	l     zerolog.Logger
}

func NewProducer(opts ...ProducerOption) (*Producer, error) {
	cfg := &ProducerConfig{
		brokers: []string{"127.0.0.1:9092"},
		topic:   "default",
		l:       zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	cli, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Producer{
		cli:   cli,
		topic: cfg.topic,
		l:     cfg.l,
	}, nil
}

func (c *Producer) Produce(ctx context.Context, msg eventstream.Streamable) (eventstream.ProduceResult, error) {
	var headers []kgo.RecordHeader
	if len(msg.StreamEventHeaders()) != 0 {
		for _, h := range msg.StreamEventHeaders() {
			headers = append(headers, kgo.RecordHeader{
				Key:   h[0],
				Value: []byte(h[1]),
			})
		}
	}

	rec := &kgo.Record{
		Key:       []byte(msg.StreamEventKey()),
		Value:     msg.StreamEventValue(),
		Headers:   headers,
		Timestamp: time.Time{},
		Topic:     c.topic,
	}

	res, err := c.cli.ProduceSync(ctx, rec).First()
	if err != nil {
		return eventstream.ProduceResult{}, fmt.Errorf("produce sync: %w", err)
	}

	c.l.Debug().
		Str("topic", res.Topic).
		Int("partition", int(res.Partition)).
		Int64("offset", res.Offset).
		Msg("kafka message written")

	return eventstream.ProduceResult{
		Partition: int(res.Partition),
		Offset:    res.Offset,
	}, nil
}

func (c *Producer) Close() {
	c.cli.Close()
}
