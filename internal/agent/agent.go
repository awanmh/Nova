package agent

import (
	"context"
)

// State defines the execution phase of the NOVA agent state machine.
type State string

const (
	StateIdle            State = "IDLE"
	StateUnderstand      State = "UNDERSTAND"
	StatePlanning        State = "PLANNING"
	StateWaitingApproval State = "WAITING_APPROVAL"
	StateExecuting       State = "EXECUTING"
	StateObserving       State = "OBSERVING"
	StateVerifying       State = "VERIFYING"
	StateReflecting      State = "REFLECTING"
	StateRecovering      State = "RECOVERING"
	StateCompleted       State = "COMPLETED"
	StateFailed          State = "FAILED"
	StateCancelled       State = "CANCELLED"
)

// Event represents a state transition or milestone during agent execution.
type Event struct {
	State   State  `json:"state"`
	Message string `json:"message"`
}

// Agent defines the interactive agent loop and state machine runtime.
type Agent interface {
	State() State
	Run(ctx context.Context, prompt string) error
	Step(ctx context.Context) (State, error)
	Cancel()
}
