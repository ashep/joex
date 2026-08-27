package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/status"
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
	now   func() time.Time
	l     zerolog.Logger
}

type ddbJob struct {
	Job
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Model string `dynamodbav:"model"`
}

type ddbTask struct {
	Task
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Model string `dynamodbav:"model"`
}

func NewRepo(ddb *dynamodb.Client, table string, now func() time.Time, l zerolog.Logger) (*Repo, error) {
	if ddb == nil {
		return nil, apperr.NewRequiredArg("ddb")
	}

	if table == "" {
		return nil, apperr.NewRequiredArg("table")
	}

	return &Repo{
		table: table,
		ddb:   ddb,
		now:   now,
		l:     l,
	}, nil
}

func (r *Repo) NewTx() *ddbtx.Tx {
	return ddbtx.New(r.ddb, r.table)
}

func (r *Repo) CreateJob(tx *ddbtx.Tx, job Job) error {
	if err := job.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	err := tx.Put(&ddbJob{
		Job:   job,
		PK:    "PIPELINE#" + job.PipelineID,
		SK:    "JOB#" + job.ID,
		Model: "job",
	}, ddbtx.WithPutCondExpr("attribute_not_exists(PK) AND attribute_not_exists(SK)"))
	if err != nil {
		return fmt.Errorf("tx put: %w", err)
	}

	return nil
}

func (r *Repo) FindJobByID(ctx context.Context, pipeID, jobID string) (Job, error) {
	qRes, err := r.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(r.table),
		ExpressionAttributeNames: map[string]string{
			"#PK": "PK",
			"#SK": "SK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PIPELINE#" + pipeID},
			":sk": &types.AttributeValueMemberS{Value: "JOB#" + jobID},
		},
		KeyConditionExpression: aws.String("#PK=:pk AND #SK=:sk"),
	})
	if err != nil {
		return Job{}, fmt.Errorf("query: %w", err)
	}

	if len(qRes.Items) == 0 {
		return Job{}, apperr.NotFoundError{Subj: jobID}
	}

	item := qRes.Items[0]
	var job ddbJob
	if err := attributevalue.UnmarshalMap(item, &job); err != nil {
		return Job{}, fmt.Errorf("unmarshal: %w", err)
	}

	return job.Job, nil
}

// FindJobsByStatus find jobs having a certain status. Limit applies to every status, not for the entire result.
func (r *Repo) FindJobsByStatus(ctx context.Context, statuses []status.ProcessingStatus, limit int) ([]Job, error) {
	if limit <= 0 {
		return nil, apperr.NewInvalidArg("limit", "must be positive")
	}

	res := make([]Job, 0)
	for _, st := range statuses {
		qRes, err := r.ddb.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(r.table),
			IndexName: aws.String("status"),
			Limit:     aws.Int32(int32(limit)),
			ExpressionAttributeNames: map[string]string{
				"#model":  "model",
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":model":  &types.AttributeValueMemberS{Value: "job"},
				":status": &types.AttributeValueMemberS{Value: string(st)},
			},
			KeyConditionExpression: aws.String("#model=:model AND #status=:status"),
		})
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}

		for _, qResItem := range qRes.Items {
			ddbItem := ddbJob{}
			if err := attributevalue.UnmarshalMap(qResItem, &ddbItem); err != nil {
				return nil, fmt.Errorf("unmarshal: %w", err)
			}
			res = append(res, ddbItem.Job)
		}
	}

	return res, nil
}

func (r *Repo) UpdateJob(tx *ddbtx.Tx, j Job) error {
	if err := j.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	expVersion := j.Version
	j.Version = j.Version + 1

	err := tx.Put(&ddbJob{
		Job:   j,
		PK:    "PIPELINE#" + j.PipelineID,
		SK:    "JOB#" + j.ID,
		Model: "job",
	},
		ddbtx.WithPutCondExpr("version=:version"),
		ddbtx.WithPutExprAttrValues(map[string]types.AttributeValue{
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(expVersion)},
		}),
	)
	if err != nil {
		return fmt.Errorf("tx put: %w", err)
	}

	return nil
}

func (r *Repo) CreateTask(tx *ddbtx.Tx, task Task) error {
	return tx.Put(&ddbTask{
		Task:  task,
		PK:    "JOB#" + task.JobID,
		SK:    "TASK#" + task.ID,
		Model: "task",
	}, ddbtx.WithPutCondExpr("attribute_not_exists(PK) AND attribute_not_exists(SK)"))
}

func (r *Repo) UpdateTask(tx *ddbtx.Tx, t Task) error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	expVersion := t.Version
	t.Version = t.Version + 1

	err := tx.Put(&ddbTask{
		Task:  t,
		PK:    "JOB#" + t.JobID,
		SK:    "TASK#" + t.ID,
		Model: "task",
	},
		ddbtx.WithPutCondExpr("version=:version"),
		ddbtx.WithPutExprAttrValues(map[string]types.AttributeValue{
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(expVersion)},
		}),
	)
	if err != nil {
		return fmt.Errorf("tx put: %w", err)
	}

	return nil
}

func (s *Repo) FindTaskByID(ctx context.Context, jobID, taskID string) (Task, error) {
	qRes, err := s.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(s.table),
		ExpressionAttributeNames: map[string]string{
			"#PK": "PK",
			"#SK": "SK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "JOB#" + jobID},
			":sk": &types.AttributeValueMemberS{Value: "TASK#" + taskID},
		},
		KeyConditionExpression: aws.String("#PK=:pk AND #SK=:sk"),
	})
	if err != nil {
		return Task{}, fmt.Errorf("query: %w", err)
	}

	if len(qRes.Items) == 0 {
		return Task{}, apperr.NotFoundError{Subj: "task " + taskID}
	}

	ddbItem := ddbTask{}
	if err := attributevalue.UnmarshalMap(qRes.Items[0], &ddbItem); err != nil {
		return Task{}, fmt.Errorf("unmarshal: %w", err)
	}

	return ddbItem.Task, nil
}

func (s *Repo) FindTasksByStatus(ctx context.Context, status status.ProcessingStatus, limit int) ([]Task, error) {
	if limit <= 0 {
		return nil, apperr.NewInvalidArg("limit", "must be positive")
	}

	qRes, err := s.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(s.table),
		IndexName: aws.String("status"),
		ExpressionAttributeNames: map[string]string{
			"#model":  "model",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":model":  &types.AttributeValueMemberS{Value: "task"},
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
		KeyConditionExpression: aws.String("#model = :model AND #status = :status"),
	})
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	res := make([]Task, len(qRes.Items))
	for i, qResItem := range qRes.Items {
		ddbItem := ddbTask{}
		if err := attributevalue.UnmarshalMap(qResItem, &ddbItem); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		res[i] = ddbItem.Task
	}

	return res, nil
}
