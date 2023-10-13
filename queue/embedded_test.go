package queue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/queue"
)

func TestNewEmbedded(t *testing.T) {
	cfg := queue.NewEmbeddedConfig()
	buff := queue.NewEmbedded(cfg)
	tasks := []boltzmann.Task{
		{
			TaskID: "0",
		},
		{
			TaskID: "1",
		},
		{
			TaskID: "2",
		},
	}
	for _, task := range tasks {
		_ = buff.Push(nil, task)
	}

	t.Log("some stuff")

	tasksOut, _ := buff.Pop(nil)
	t.Log(len(tasksOut))
	assert.Equal(t, len(tasks), len(tasksOut))
}
