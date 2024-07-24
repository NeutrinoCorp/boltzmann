package state

import (
	"fmt"
	"time"

	"github.com/modern-go/reflect2"
)

type State struct {
	StateID           string
	ResourceName      string
	ResourceID        string
	Status            string
	StartTime         time.Time
	EndTime           time.Time
	ExecutionDuration time.Duration
	ExecutionError    string
}

func newStateID[T any](resourceID string) string {
	var zeroVal T
	typeOf := reflect2.TypeOf(zeroVal).String()
	return fmt.Sprintf("%s#%s", typeOf, resourceID)
}

func createState[T any](resourceID string) State {
	var zeroVal T
	typeOf := reflect2.TypeOf(zeroVal).String()
	return State{
		StateID:      newStateID[T](resourceID),
		ResourceName: typeOf,
		ResourceID:   resourceID,
		Status:       StatusRunning,
		StartTime:    time.Now().UTC(),
	}
}

func (s *State) markWithState(status string, procErr error) {
	if procErr != nil {
		s.ExecutionError = procErr.Error()
	}
	s.Status = status
	s.EndTime = time.Now().UTC()
	s.ExecutionDuration = s.EndTime.Sub(s.StartTime)
}

func (s *State) markAsFailed(procErr error) {
	s.markWithState(StatusFailed, procErr)
}

func (s *State) markAsCompleted() {
	s.markWithState(StatusCompleted, nil)
}

func (s *State) View() View {
	return View{
		StateID:                 s.StateID,
		ResourceName:            s.ResourceName,
		ResourceID:              s.ResourceID,
		Status:                  s.Status,
		StartTime:               s.StartTime,
		StartTimeMillis:         s.StartTime.UnixMilli(),
		EndTime:                 s.EndTime,
		EndTimeMillis:           s.EndTime.UnixMilli(),
		ExecutionDuration:       s.ExecutionDuration,
		ExecutionDurationMillis: s.ExecutionDuration.Milliseconds(),
		ExecutionError:          s.ExecutionError,
	}
}
