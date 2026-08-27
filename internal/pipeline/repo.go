package pipeline

import (
	"context"
	"fmt"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/pkg/ddbtx"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

type Repo struct {
	table string
	ddb   *dynamodb.Client
	l     zerolog.Logger
}

type ddbPipeline struct {
	Pipeline
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Model string `dynamodbav:"model"`
}

func NewRepo(ddb *dynamodb.Client, table string, l zerolog.Logger) (*Repo, error) {
	if ddb == nil {
		return nil, apperr.NewRequiredArg("ddb")
	}

	if table == "" {
		return nil, apperr.NewRequiredArg("table")
	}

	return &Repo{
		table: table,
		ddb:   ddb,
		l:     l,
	}, nil
}

func (r *Repo) NewTx() *ddbtx.Tx {
	return ddbtx.New(r.ddb, r.table)
}

func (s *Repo) Create(tx *ddbtx.Tx, p Pipeline) error {
	err := tx.Put(&ddbPipeline{
		Pipeline: p,
		PK:       "PIPELINE#" + p.ID,
		SK:       "PIPELINE",
		Model:    "pipeline",
	}, ddbtx.WithPutCondExpr("attribute_not_exists(PK)"))
	if err != nil {
		return fmt.Errorf("tx put: %w", err)
	}

	return nil
}

func (s *Repo) FindByID(ctx context.Context, id string) (Pipeline, error) {
	qRes, err := s.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		KeyConditionExpression: aws.String(`PK=:PK AND SK=:SK`),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: "PIPELINE#" + id},
			":SK": &types.AttributeValueMemberS{Value: "PIPELINE"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return Pipeline{}, fmt.Errorf("query: %w", err)
	}

	if len(qRes.Items) == 0 {
		return Pipeline{}, apperr.NotFoundError{Subj: "pipeline"}
	}

	var pItem ddbPipeline
	if err := attributevalue.UnmarshalMap(qRes.Items[0], &pItem); err != nil {
		return Pipeline{}, fmt.Errorf("unmarshal pipeline: %w", err)
	}

	p := Pipeline{
		ID:        pItem.ID,
		Status:    pItem.Status,
		Steps:     pItem.Steps,
		Version:   pItem.Version,
		CreatedAt: pItem.CreatedAt,
		UpdatedAt: pItem.UpdatedAt,
	}
	if err := p.Validate(); err != nil {
		return Pipeline{}, fmt.Errorf("make pipeline: %w", err)
	}

	return p, nil
}
