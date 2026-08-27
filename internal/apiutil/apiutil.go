package apiutil

import (
	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	proto "github.com/ashep/joex/sdk/proto/joex/v1"
)

func MapProtoDataTypeArgs(args map[string]*proto.Arg) (datatype.VarMap, error) {
	res := make(datatype.VarMap, len(args))
	for name, opt := range args {
		if err := res.Set(name, mapDataType(args[name].GetType()), opt.Value); err != nil {
			return nil, apperr.NewInvalidArg(name, err.Error())
		}
	}
	return res, nil
}

func mapDataType(arg proto.DataType) datatype.Type {
	switch arg {
	case proto.DataType_DATA_TYPE_BOOL:
		return datatype.Bool
	case proto.DataType_DATA_TYPE_INT:
		return datatype.Int
	case proto.DataType_DATA_TYPE_FLOAT:
		return datatype.Float
	case proto.DataType_DATA_TYPE_STRING:
		return datatype.String
	default:
		return datatype.Unknown
	}
}
