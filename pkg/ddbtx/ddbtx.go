package ddbtx

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ErrEmptyItems = errors.New("empty items")
)

type Tx struct {
	ddb       *dynamodb.Client
	tableName string
	items     []types.TransactWriteItem
}

func (tx *Tx) Put(in any, opts ...PutOption) error {
	item, err := attributevalue.MarshalMap(in)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	put := &types.Put{
		Item:      item,
		TableName: aws.String(tx.tableName),
	}

	for _, opt := range opts {
		opt(put)
	}

	tx.items = append(tx.items, types.TransactWriteItem{
		Put: put,
	})

	return nil
}

func (tx *Tx) Update(key map[string]types.AttributeValue, updExp string, opts ...UpdateOption) error {
	upd := &types.Update{
		Key:              key,
		TableName:        aws.String(tx.tableName),
		UpdateExpression: &updExp,
	}

	for _, opt := range opts {
		opt(upd)
	}

	tx.items = append(tx.items, types.TransactWriteItem{
		Update: upd,
	})

	return nil
}

func (tx *Tx) Commit(ctx context.Context) (*dynamodb.TransactWriteItemsOutput, error) {
	if len(tx.items) == 0 {
		return nil, ErrEmptyItems
	}

	req := &dynamodb.TransactWriteItemsInput{
		TransactItems: tx.items,
	}

	return tx.ddb.TransactWriteItems(ctx, req)
}

func New(ddb *dynamodb.Client, tableName string) *Tx {
	return &Tx{
		ddb:       ddb,
		tableName: tableName,
		items:     make([]types.TransactWriteItem, 0),
	}
}
