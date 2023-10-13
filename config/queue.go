package config

const (
	QueueName          = "QUEUE_NAME"           // Name of the task queue (if applicable).
	QueueBatchSize     = "QUEUE_BATCH_SIZE"     // Size of the queue (if applicable).
	QueueFetchInterval = "QUEUE_FETCH_INTERVAL" // Interval for the queuing service to fetch next task batch.
	QueueRetryInterval = "QUEUE_RETRY_INTERVAL" // Interval for the queuing service to retry a failed fetching process.
	QueueJobTimeout    = "QUEUE_JOB_TIMEOUT"    // Timeout for a single task to be completed.
	QueueMaxRetry      = "QUEUE_MAX_RETRY"      // Maximum times a queuing service might fetch a task batch after an error. If -1, then unlimited retries will be available.
	QueueMaxProc       = "QUEUE_MAX_PROC"       // Resource limit for a queuing service (max. goroutines).
)
