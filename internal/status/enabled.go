package status

import (
	apperr "github.com/ashep/go-app/errors"
)

const (
	Enabled  = "enabled"
	Disabled = "disabled"
)

type EnabledStatus string

func ValidateEnabledStatus(s EnabledStatus) error {
	switch s {
	case Enabled, Disabled:
		return nil
	default:
		return apperr.NewInvalidArg(string(s), "invalid enabled status")
	}
}
