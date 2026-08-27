package ddboutbox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/pkg/eventstream"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

var ErrNoNMoreItems = errors.New("no more items found")

type txItemProvider[T any] func(T) (*types.TransactWriteItem, error)

type streamer interface {
	Produce(ctx context.Context, msg eventstream.Streamable) (eventstream.ProduceResult, error)
}

type ddbItem struct {
	PK    string `json:"-" dynamodbav:"PK"`
	SK    string `json:"-" dynamodbav:"SK"`
	Value []byte `json:"value" dynamodbav:"value"`
}

func (d ddbItem) StreamEventKey() string {
	return d.SK
}

func (d ddbItem) StreamEventValue() []byte {
	return d.Value
}

func (d ddbItem) StreamEventHeaders() [][2]string {
	return nil
}

func (d ddbItem) StreamEventTime() time.Time {
	sk := strings.SplitN(d.SK, "_", 2)
	if len(sk) != 2 {
		return time.Time{}
	}

	ms, err := strconv.ParseInt(sk[0], 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.UnixMilli(ms)
}

type Outbox[T eventstream.Streamable] struct {
	ddb            *dynamodb.Client
	name           string
	table          string
	dbItemProvider txItemProvider[T]
	stream         streamer
	pollPeriod     time.Duration
	l              zerolog.Logger
	stopped        chan struct{}
}

func New[T eventstream.Streamable](
	ddb *dynamodb.Client,
	name string,
	table string,
	itemProvider txItemProvider[T],
	stream streamer,
	opts ...Option,
) (*Outbox[T], error) {
	if ddb == nil {
		return nil, errors.New("ddb is required")
	}

	if name == "" {
		return nil, apperr.NewRequiredArg("name")
	}

	if table == "" {
		return nil, apperr.NewRequiredArg("table name")
	}

	if itemProvider == nil {
		return nil, apperr.NewRequiredArg("txItemCreator")
	}

	if stream == nil {
		return nil, apperr.NewRequiredArg("stream")
	}

	cfg := config{
		pollPeriod: time.Second,
		l:          zerolog.Nop(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Outbox[T]{
		ddb:            ddb,
		name:           name,
		table:          table,
		dbItemProvider: itemProvider,
		stream:         stream,
		pollPeriod:     cfg.pollPeriod,
		l:              cfg.l,
		stopped:        make(chan struct{}),
	}, nil
}

// Send stores the item in the outbox and writes the associated data item in a single DynamoDB transaction.
func (o *Outbox[T]) Send(ctx context.Context, item T) error {
	outboxItem, err := attributevalue.MarshalMap(&ddbItem{
		PK:    "OUTBOX_" + o.name,
		SK:    strconv.Itoa(int(item.StreamEventTime().UnixMilli())) + "_" + item.StreamEventKey(),
		Value: item.StreamEventValue(),
	})
	if err != nil {
		return fmt.Errorf("marshal outbox item: %w", err)
	}

	dataItem, err := o.dbItemProvider(item)
	if err != nil {
		return fmt.Errorf("marshal data item: %w", err)
	}

	txItems := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName: aws.String(o.table),
				Item:      outboxItem,
			},
		},
		*dataItem,
	}

	_, err = o.ddb.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: txItems,
	})
	if err != nil {
		return fmt.Errorf("db: write items: %w", err)
	}

	return nil
}

// Run continuously polls the outbox for items, produces them to the stream, and deletes them upon success.
func (o *Outbox[T]) Run(ctx context.Context) error {
	defer close(o.stopped)

	sleep := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		sleep = o.pollPeriod

		item, err := o.nextItem(ctx)
		if errors.Is(err, ErrNoNMoreItems) {
			continue
		} else if err != nil {
			o.l.Error().Err(err).Msg("failed to get next item")
			continue
		}

		if _, err := o.stream.Produce(ctx, item); err != nil {
			o.l.Error().Err(err).Msg("failed to produce stream event")
			continue
		}
		o.l.Debug().Str("key", item.SK).Msg("item streamed")

		if err := o.deleteItem(ctx, item.SK); err != nil {
			o.l.Error().Err(err).Str("key", item.SK).Msg("failed to delete item")
			continue
		}
		o.l.Debug().Str("key", item.SK).Msg("item deleted")

		sleep = 0
	}
}

func (o *Outbox[T]) nextItem(ctx context.Context) (ddbItem, error) {
	out, err := o.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(o.table),
		KeyConditionExpression: aws.String("#PK=:pk"),
		ExpressionAttributeNames: map[string]string{
			"#PK": "PK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OUTBOX_" + o.name},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return ddbItem{}, fmt.Errorf("scan: %w", err)
	}

	if len(out.Items) == 0 {
		return ddbItem{}, ErrNoNMoreItems
	}

	res := ddbItem{}
	if err := attributevalue.UnmarshalMap(out.Items[0], &res); err != nil {
		return res, fmt.Errorf("unmarshal db item: %w", err)
	}

	return res, nil
}

func (o *Outbox[T]) deleteItem(ctx context.Context, key string) error {
	_, err := o.ddb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(o.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OUTBOX_" + o.name},
			"SK": &types.AttributeValueMemberS{Value: key},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	})

	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return apperr.NotFoundError{Subj: "item"}
		}

		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (o *Outbox[T]) Wait() error {
	<-o.stopped
	return nil
}
