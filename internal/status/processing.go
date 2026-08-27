package status

import (
	apperr "github.com/ashep/go-app/errors"
)

const (
	New     ProcessingStatus = "new"
	Idle    ProcessingStatus = "idle"
	Running ProcessingStatus = "running"
	Done    ProcessingStatus = "done"
	Failed  ProcessingStatus = "failed"
)

type ProcessingStatus string

func ValidateProcessingStatus(s ProcessingStatus) error {
	switch s {
	case New, Idle, Running, Done, Failed:
		return nil
	default:
		return apperr.NewInvalidArg(string(s), "invalid processing status")
	}
}
