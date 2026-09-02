package apiutil

import (
	"fmt"
	"strings"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"google.golang.org/protobuf/types/known/structpb"
)

func MapProtoStructToDataType(args *structpb.Struct) (datatype.VarMap, error) {
	fields := args.GetFields()
	res := make(datatype.VarMap, len(fields))
	for k, v := range fields {
		var err error

		k = strings.TrimSpace(k)
		if k == "" {
			return nil, apperr.NewRequiredArg("key")
		}

		switch kT := v.GetKind().(type) {
		case *structpb.Value_BoolValue:
			err = res.SetBool(k, fmt.Sprintf("%v", v.GetBoolValue()))
		case *structpb.Value_NumberValue:
			err = res.SetNumber(k, fmt.Sprintf("%v", v.GetNumberValue()))
		case *structpb.Value_StringValue:
			err = res.SetString(k, v.GetStringValue())
		default:
			err = apperr.NewInvalidArg(k, fmt.Sprintf("unexpected type: %T", kT))
		}

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
