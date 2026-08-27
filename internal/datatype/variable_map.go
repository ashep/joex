package datatype

import (
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
	case Int:
		v, err = NewInt(val)
	case Float:
		v, err = NewFloat(val)
	case String:
		v = NewString(val)
	default:
		err = apperr.NewInvalidArg("typ", "unexpected type")
	}

	if err != nil {
		return err
	}

	m[key] = v

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

func (m VarMap) SetInt(key string, val string) error {
	return m.Set(key, Int, val)
}

func (m VarMap) Int(key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, &NotDefinedError{Name: key}
	}
	return v.Int()
}

func (m VarMap) SetFloat(key string, val string) error {
	return m.Set(key, Float, val)
}

func (m VarMap) Float(key string) (float64, error) {
	v, ok := m[key]
	if !ok {
		return 0, &NotDefinedError{Name: key}
	}
	return v.Float()
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
