package queue

import (
	"context"

	"github.com/eapache/queue/v2"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/config"
)

type EmbeddedConfig struct {
	BufferSize int
	BatchSize  int
}

func setEmbeddedConfigDefault() {
	config.SetDefault(config.EmbeddedQueueSize, 100)
	config.SetDefault(config.QueueBatchSize, 25)
}

func NewEmbeddedConfig() EmbeddedConfig {
	setEmbeddedConfigDefault()
	return EmbeddedConfig{
		BufferSize: config.Get[int](config.EmbeddedQueueSize),
		BatchSize:  config.Get[int](config.QueueBatchSize),
	}
}

// Embedded is the Queue implementation using an in-memory circular buffer.
type Embedded struct {
	Config EmbeddedConfig
	r      *queue.Queue[boltzmann.Task]
}

func NewEmbedded(cfg EmbeddedConfig) *Embedded {
	return &Embedded{
		Config: cfg,
		r:      queue.New[boltzmann.Task](),
	}
}

var _ Queue = &Embedded{}

func (e *Embedded) Push(_ context.Context, tasks ...boltzmann.Task) error {
	for _, task := range tasks {
		e.r.Add(task)
	}
	return nil
}

func (e *Embedded) Pop(_ context.Context) ([]boltzmann.Task, error) {
	if e.r.Length() == 0 {
		return nil, nil
	}

	tasks := make([]boltzmann.Task, 0, e.Config.BatchSize)
	tmpLen := e.r.Length()
	for i := 0; i < tmpLen; i++ {
		if i == e.Config.BatchSize {
			break
		}

		tasks = append(tasks, e.r.Remove())
	}
	return tasks, nil
}
