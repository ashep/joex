package typeutil

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	_ attributevalue.Marshaler   = UnixTimeMs{}
	_ attributevalue.Unmarshaler = &UnixTimeMs{}
)

type UnixTimeMs time.Time

func (t UnixTimeMs) IsZero() bool {
	return t.AsTime().IsZero()
}

func (t UnixTimeMs) AsTime() time.Time {
	return time.Time(t)
}

func (t UnixTimeMs) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{
		Value: strconv.FormatInt(t.AsTime().UnixMilli(), 10),
	}, nil
}

func (t *UnixTimeMs) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	n, ok := av.(*types.AttributeValueMemberN)
	if !ok {
		return fmt.Errorf("unexpected type %T", av)
	}
	ms, err := strconv.ParseInt(n.Value, 10, 64)
	if err != nil {
		return err
	}

	*t = UnixTimeMs(time.UnixMilli(ms))

	return nil
}

func (t UnixTimeMs) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(t.AsTime().UnixMilli()))), nil
}
