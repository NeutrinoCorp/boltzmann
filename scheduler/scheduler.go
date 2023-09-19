package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/state"
)

type ScheduleTaskResult struct {
	TaskID        string
	CorrelationID string
	Driver        string
	ResourceURI   string
	ErrorMessage  string
	ScheduleTime  time.Time
}

type TaskScheduler interface {
	Schedule(ctx context.Context, tasks []boltzmann.Task) []ScheduleTaskResult
}

type SyncTaskScheduler struct {
	AgentRegistry   agent.Registry
	QueueService    queue.Service
	StateRepository state.Repository
}

var _ TaskScheduler = SyncTaskScheduler{}

func (s SyncTaskScheduler) Schedule(ctx context.Context, tasks []boltzmann.Task) []ScheduleTaskResult {
	wg := sync.WaitGroup{}
	wg.Add(len(tasks))

	resultsAtomic := atomic.Value{}
	resultsAtomic.Store(make([]ScheduleTaskResult, 0))
	for _, task := range tasks {
		go func(taskCopy boltzmann.Task, startTime time.Time) {
			defer wg.Done()
			var err error
			_, err = s.AgentRegistry.Get(taskCopy.Driver)
			if err != nil {
				resultsAtomic.Store(append(resultsAtomic.Load().([]ScheduleTaskResult), ScheduleTaskResult{
					TaskID:        taskCopy.TaskID,
					CorrelationID: taskCopy.CorrelationID,
					Driver:        taskCopy.Driver,
					ResourceURI:   taskCopy.ResourceURI,
					ErrorMessage:  err.Error(),
					ScheduleTime:  startTime,
				}))
				return
			}

			scopedCtx, cancel := context.WithTimeout(ctx, time.Second*60)
			defer cancel()
			defer func() {
				taskCopy.Status = boltzmann.TaskStatusSucceed
				if err != nil {
					taskCopy.Status = boltzmann.TaskStatusFailed
					taskCopy.FailureMessage = err.Error()
					resultsAtomic.Store(append(resultsAtomic.Load().([]ScheduleTaskResult), ScheduleTaskResult{
						TaskID:        taskCopy.TaskID,
						CorrelationID: taskCopy.CorrelationID,
						Driver:        taskCopy.Driver,
						ResourceURI:   taskCopy.ResourceURI,
						ErrorMessage:  err.Error(),
						ScheduleTime:  startTime,
					}))
				}

				if errSave := s.StateRepository.Save(scopedCtx, taskCopy); errSave != nil {
					log.Err(err).
						Str("task_id", taskCopy.TaskID).
						Str("driver", taskCopy.Driver).
						Str("resource_location", taskCopy.ResourceURI).
						Msg("failed to save state")
				}
			}()
			errSave := s.StateRepository.Save(scopedCtx, taskCopy)
			if errSave != nil {
				err = errSave
				return
			} // commit state

			taskCopy.Status = boltzmann.TaskStatusScheduled
			taskCopy.StartTime = startTime
			err = s.QueueService.Enqueue(scopedCtx, taskCopy)
			if err != nil {
				log.Err(err).
					Str("task_id", taskCopy.TaskID).
					Str("driver", taskCopy.Driver).
					Str("resource_location", taskCopy.ResourceURI).
					Msg("cannot enqueue task")
				return
			}

			errSave = s.StateRepository.Save(scopedCtx, taskCopy)
			if errSave != nil {
				err = errSave
				return
			}
			log.Info().
				Str("task_id", taskCopy.TaskID).
				Str("driver", taskCopy.Driver).
				Str("resource_location", taskCopy.ResourceURI).
				Msg("successfully scheduled task")
			resultsAtomic.Store(append(resultsAtomic.Load().([]ScheduleTaskResult), ScheduleTaskResult{
				TaskID:        taskCopy.TaskID,
				CorrelationID: taskCopy.CorrelationID,
				Driver:        taskCopy.Driver,
				ResourceURI:   taskCopy.ResourceURI,
				ScheduleTime:  startTime,
			}))
		}(task, time.Now().UTC())
	}
	wg.Wait()

	return resultsAtomic.Load().([]ScheduleTaskResult)
}
