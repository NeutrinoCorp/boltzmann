package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/codec"
)

type Service struct {
	StateRepository Repository
	Codec           codec.Codec
}

func (s Service) Get(ctx context.Context, taskID string) (boltzmann.Task, error) {
	return s.StateRepository.Get(ctx, taskID)
}

// GetIfChanged retrieves a task state only if state changed. Returns task using codec.Codec, current hash or an error.
func (s Service) GetIfChanged(ctx context.Context, prevHash string, taskID string) (boltzmann.Task, string, error) {
	// Having a global hashFunc instance, so it may be reused WILL NOT work in concurrent scenarios (not thread-safe).
	hashFunc := sha256.New()
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return boltzmann.Task{}, "", err
	}

	// Maybe we could use another way to calculate task hash instead marshaling
	// DO NOT use jsoniter package as it leads to unexpected behaviour, causing hash mismatching.
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return boltzmann.Task{}, "", err
	}

	_, _ = io.Copy(hashFunc, bytes.NewReader(taskJSON))
	currentHash := hex.EncodeToString(hashFunc.Sum(nil))
	if currentHash == prevHash {
		return boltzmann.Task{}, prevHash, nil
	}

	return task, currentHash, nil
}
