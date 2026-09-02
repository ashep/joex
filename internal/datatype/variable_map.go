package datatype

import (
	"fmt"

	apperr "github.com/ashep/go-app/errors"
)

type VarMap map[string]Var

func (m VarMap) Set(key string, typ Type, val string) error {
	v, ok := m[key]
	if ok && v.Type != typ {
		return apperr.NewInvalidArg(key, "is not "+string(typ))
	}

	var (
		err error
	)
	switch typ {
	case Bool:
		v, err = NewBool(val)
		if err != nil {
			err = apperr.NewInvalidArg(key, err.Error())
		}
	case Number:
		v, err = NewNumber(val)
		if err != nil {
			err = apperr.NewInvalidArg(key, err.Error())
		}
	case String:
		v = NewString(val)
	default:
		err = apperr.NewInvalidArg(key, "unexpected type: "+string(typ))
	}

	if err != nil {
		return err
	}

	m[key] = v

	return nil
}

func (m VarMap) AsMap() map[string]any {
	res := make(map[string]any, len(m))
	for k, v := range m {
		res[k] = v.Value
	}
	return res
}

func (m VarMap) FromMap(in map[string]any) error {
	for k, v := range in {
		var err error

		switch v.(type) {
		case bool:
			err = m.SetBool(k, fmt.Sprintf("%v", v))
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			err = m.SetNumber(k, fmt.Sprintf("%v", v))
		case string:
			err = m.SetString(k, fmt.Sprintf("%v", v))
		default:
			err = apperr.NewInvalidArg(k, fmt.Sprintf("unsupported type: %T", v))
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (m VarMap) SetBool(key string, val string) error {
	return m.Set(key, Bool, val)
}

func (m VarMap) Bool(key string) (bool, error) {
	v, ok := m[key]
	if !ok {
		return false, &NotDefinedError{Name: key}
	}
	return v.Bool()
}

func (m VarMap) MustBool(key string) bool {
	v, _ := m.Bool(key)
	return v
}

func (m VarMap) SetNumber(key string, val string) error {
	return m.Set(key, Number, val)
}

func (m VarMap) Number(key string) (float64, error) {
	v, ok := m[key]
	if !ok {
		return 0, &NotDefinedError{Name: key}
	}
	return v.Number()
}

func (m VarMap) MustNumber(key string) float64 {
	v, _ := m.Number(key)
	return v
}

func (m VarMap) SetString(key string, val string) error {
	return m.Set(key, String, val)
}

func (m VarMap) String(key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", &NotDefinedError{Name: key}
	}
	return v.String()
}

func (m VarMap) MustString(key string) string {
	v, _ := m.String(key)
	return v
}
