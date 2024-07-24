package agent

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Agent interface {
	ExecTask(ctx context.Context, task boltzmann.Task) error
}

// TODO: Add actual HTTP agent
//  - Maybe add http.headers argument and try to decode a map[string]string. This would keep agent args clean.
//    Or use a `header.` prefix to detect header args in driver.
