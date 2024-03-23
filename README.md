# Boltzmann
Boltzmann is a distributed lightweight arg orchestrator.

Based on the [Scheduler Agent Supervisor Cloud Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/scheduler-agent-supervisor),
`Boltzmann` is a master-less service used to schedule a batch of arg in a parallel and distributed way.

Depending on the configuration, a `Boltzmann` node might be stateless or stateful as args states may be stored in a 
embedded or external database (_e.g. Redis_).

Worker pools (_i.e. a `Boltzmann` node_) are ensured for correctness even in a distributed environment by using
[leases](https://martinfowler.com/articles/patterns-of-distributed-systems/time-bound-lease.html) (_i.e. distributed 
mutex lock_) and a small [leader election consensus algorithm](https://aws.amazon.com/builders-library/leader-election-in-distributed-systems/).

Moreover, `Leases` are implemented using either a _RedLock algorithm_ or through storage engine's built-in data structure
(_e.g. etcd leases_).

## Architecture

![High-Level Archictecture Diagram](https://learn.microsoft.com/en-us/azure/architecture/patterns/_images/scheduler-agent-supervisor-pattern.png)

### Task Scheduler

The `Scheduler` arranges for the steps that make up the arg to be executed and orchestrates their operation. These steps 
can be combined into a pipeline or workflow. The Scheduler is responsible for ensuring that the steps in this workflow 
are performed in the right order.

As each step is performed, the Scheduler records the state of the workflow, such as 
"step not yet started," "step running," or "step completed." The state information should also include an upper limit 
of the time allowed for the step to finish, called the complete-by time.

If a step requires access to a remote service or resource, the Scheduler invokes the appropriate Agent, passing it the 
details of the work to be performed. The Scheduler typically communicates with an Agent using asynchronous request/response messaging.

### Agent

The `Agent` contains logic that encapsulates a call to a remote service, or access to a remote resource referenced by a 
step in a arg. Each Agent typically wraps calls to a single service or resource, implementing the appropriate error 
handling and retry logic (subject to a timeout constraint, described later).

### Supervisor

The Supervisor monitors the status of the steps in the arg being performed by the Scheduler. It runs periodically 
(the frequency will be system-specific), and examines the status of steps maintained by the Scheduler. If it detects 
any that have timed out or failed, it arranges for the appropriate Agent to recover the step or execute the appropriate 
remedial action (this might involve modifying the status of a step).

Note that the recovery or remedial actions are implemented by the Scheduler and Agents. The Supervisor should simply 
request that these actions be performed.

## Usage

Till this day, there are two ways available to use `Boltzmann` (_which are not mutually exclusive_):

- A HTTP REST API (_HTTP/1.1_).
- A gRCP Streaming API (_HTTP/2, multiplexed_).
