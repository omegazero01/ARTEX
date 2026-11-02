package agent

import "context"

// IntentStopResult is what a user-initiated intent cancellation reports back.
type IntentStopResult struct {
	ID              int64  `json:"id"`
	State           string `json:"state"`
	Reason          string `json:"reason,omitempty"`
	CancelledByUser bool   `json:"cancelled_by_user"`
}

// IntentStopper stops one intent of the current task on behalf of the user.
type IntentStopper func(ctx context.Context, taskID, intentID int64, reason string) (IntentStopResult, error)

// SetCancelIntent wires the server-side user-cancel path into the main agent.
func (m *MainAgent) SetCancelIntent(fn IntentStopper) { m.cancelIntent = fn }
