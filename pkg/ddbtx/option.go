package ddbtx

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type PutOption func(put *types.Put)

func WithPutCondExpr(
	expr string,
) PutOption {
	return func(put *types.Put) {
		put.ConditionExpression = &expr
	}
}

func WithPutExprAttrNames(attrNames map[string]string) PutOption {
	return func(put *types.Put) {
		put.ExpressionAttributeNames = attrNames
	}
}

func WithPutExprAttrValues(attrValues map[string]types.AttributeValue) PutOption {
	return func(put *types.Put) {
		put.ExpressionAttributeValues = attrValues
	}
}

type UpdateOption func(upd *types.Update)

func WithUpdateCondExpr(expr string) UpdateOption {
	return func(upd *types.Update) {
		upd.ConditionExpression = &expr
	}
}

func WithUpdateExprAttrNames(attrNames map[string]string) UpdateOption {
	return func(upd *types.Update) {
		upd.ExpressionAttributeNames = attrNames
	}
}

func WithUpdateExprAttrValues(attrValues map[string]types.AttributeValue) UpdateOption {
	return func(upd *types.Update) {
		upd.ExpressionAttributeValues = attrValues
	}
}
