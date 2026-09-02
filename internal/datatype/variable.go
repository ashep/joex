package datatype

import (
	"fmt"
	"strconv"
	"strings"

	apperr "github.com/ashep/go-app/errors"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TypeError struct {
	Expected Type
	Actual   Type
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("is not %s, but %s", e.Expected, e.Actual)
}

type NotDefinedError struct {
	Name string
}

func (e *NotDefinedError) Error() string {
	return fmt.Sprintf("%s is not defined", e.Name)
}

type Type string

const (
	Unspecified Type = "unspecified"
	Bool        Type = "bool"
	Number      Type = "number"
	String      Type = "string"
)

type Var struct {
	Type  Type `json:"type"`
	Value any  `json:"value"`
}

func (v Var) Bool() (bool, error) {
	if v.Type != Bool {
		return false, &TypeError{Expected: Bool, Actual: v.Type}
	}
	return v.Value.(bool), nil
}

func (v Var) Number() (float64, error) {
	if v.Type != Number {
		return 0, &TypeError{Expected: Number, Actual: v.Type}
	}
	return v.Value.(float64), nil
}

func (v Var) String() (string, error) {
	if v.Type != String {
		return "", &TypeError{Expected: String, Actual: v.Type}
	}
	return v.Value.(string), nil
}

func (v Var) AsString() string {
	switch v.Type {
	case Bool:
		return fmt.Sprintf("%t", v.Value.(bool))
	case Number:
		return fmt.Sprintf("%f", v.Value.(float64))
	case String:
		return v.Value.(string)
	default:
		return ""
	}
}

func NewBool(s string) (Var, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "true":
		return Var{Type: Bool, Value: true}, nil
	case "false":
		return Var{Type: Bool, Value: false}, nil
	default:
		return Var{}, apperr.NewInvalidArg(s, "cannot be converted to a bool")
	}
}

func NewNumber(s string) (Var, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(strings.ToLower(s)), 64)
	if err != nil {
		return Var{}, apperr.NewInvalidArg(s, fmt.Sprintf("cannot be converted to a float: %s", err))
	}
	return Var{Type: Number, Value: f}, nil
}

func NewString(s string) Var {
	return Var{Type: String, Value: s}
}

func (v *Var) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	m, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("expected AttributeValueMemberM, got %T", av)
	}

	typeAV, ok := m.Value["Type"]
	if !ok {
		return fmt.Errorf("missing field: Type")
	}
	if err := attributevalue.Unmarshal(typeAV, &v.Type); err != nil {
		return fmt.Errorf("unmarshal type: %w", err)
	}

	valueAV, ok := m.Value["Value"]
	if !ok {
		return fmt.Errorf("missing field: Value")
	}

	switch v.Type {
	case Bool:
		var b bool
		if err := attributevalue.Unmarshal(valueAV, &b); err != nil {
			return fmt.Errorf("unmarshal bool: %w", err)
		}
		v.Value = b
	case Number:
		var f float64
		if err := attributevalue.Unmarshal(valueAV, &f); err != nil {
			return fmt.Errorf("unmarshal number: %w", err)
		}
		v.Value = f
	case String:
		var s string
		if err := attributevalue.Unmarshal(valueAV, &s); err != nil {
			return fmt.Errorf("unmarshal string: %w", err)
		}
		v.Value = s
	default:
		return fmt.Errorf("unknown type: %q", v.Type)
	}

	return nil
}
