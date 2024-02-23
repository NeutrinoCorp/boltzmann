package boltzmann

import "context"

type BackgroundProcess interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
