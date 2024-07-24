package supervisor

// TODO: Add several failure types (e.g. RETRYABLE,FATAL). Aids in failing fast, avoids poison messages.
// TODO: Add RetryFairTask() in scheduler service API. Use this from supervisor to retry a task so the sched can detect exec plan
//  and execute tasks from the last succeeded step (if fairness=ON).
// TODO: Add RetryTask() in sched service API to retry task within an exec plan with fairness=OFF.
