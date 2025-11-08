#  DESIGN.md

##  Overview

`queuectl` is a **CLI-based background job queue system** built in Go.
It manages background jobs with persistence, concurrency, retries, and monitoring — all via a simple command-line interface.

It is designed for **local or small-scale production use**, focusing on reliability, extensibility, and clarity.

---

##  Goals

* Reliable background job processing
* Persistent storage using SQLite
* Automatic retries with exponential backoff
* Dead Letter Queue (DLQ) for failed jobs
* Configurable job priorities and scheduling (`run_at`, `--delay`)
* Graceful shutdown and timeout handling
* Extensible CLI for easy management

---

##  Core Components

### 1. **Job Model**

Each job is a persistent record in SQLite.
This model defines the entire job lifecycle and its metadata.

| Field                       | Type               | Description                                                               |
| --------------------------- | ------------------ | ------------------------------------------------------------------------- |
| `id`                        | string             | Unique job identifier                                                     |
| `command`                   | string             | Shell command to execute                                                  |
| `state`                     | string             | Job state — one of `pending`, `processing`, `completed`, `failed`, `dead` |
| `attempts`                  | int                | Number of attempts made                                                   |
| `max_retries`               | int                | Maximum allowed retries                                                   |
| `priority`                  | int                | Determines execution order (higher = earlier)                             |
| `run_at`                    | datetime           | Scheduled execution time (for delayed jobs)                               |
| `duration`                  | float              | Execution time in seconds                                                 |
| `output`                    | text               | Captured command output (stdout/stderr)                                   |
| `last_error`                | text               | Error message from last failure                                           |
| `created_at` / `updated_at` | timestamps         | Lifecycle tracking                                                        |
| `deleted_at`                | nullable timestamp | Soft delete via GORM                                                      |

---

### 2. **Persistent Storage (`internal/store`)**

Implements a repository layer over **SQLite via GORM**.

**Responsibilities:**

* Insert new jobs (`Create`)
* Fetch pending jobs with priority and scheduling awareness (`FindPending`)
* Safely claim jobs with transactional locking (`PreventRaceCondition`)
* Update job states (`Processing`, `MarkCompleted`, `Failed`)
* Move exhausted jobs to the **Dead Letter Queue**
* Gather queue metrics (`JobMetrics`)

**Features:**

* Exponential retry backoff → `delay = baseDelay * 2^(attempts-1)`
* Transaction-based job claiming prevents duplicate processing
* Priority-aware job ordering → higher-priority jobs picked first
* Aggregated metrics for reporting and dashboarding

---

### 3. **Worker System (`internal/queue`)**

Each worker runs as a **goroutine**, polling the database for available jobs and executing them safely.

**Responsibilities:**

* Poll for pending jobs (`FindPending` / `PreventRaceCondition`)
* Execute shell commands using `exec.CommandContext` (supports timeout)
* Capture output and exit code
* Update state to `completed`, `failed`, or `dead`
* Respect exponential backoff and retry logic
* Graceful shutdown on interrupt signals

**Worker Configurable Flags:**

* `--count` → Number of concurrent workers
* `--timeout` → Max runtime per job
* `--backoff-base` → Base delay for exponential backoff

---

### 4. **CLI Layer (`cmd/`)**

Built with **Cobra**, the CLI provides clean, intuitive commands for managing jobs and workers.

| Command        | Description                                                                 |
| -------------- | --------------------------------------------------------------------------- |
| `enqueue`      | Add a new job to the queue. Supports `--priority` and `--delay`.            |
| `worker start` | Start worker(s) with optional `--count`, `--timeout`, and `--backoff-base`. |
| `list`         | List all jobs by state. Optionally show command output.                     |
| `stats`        | View aggregated metrics like totals, averages, and retry counts.            |
| `dlq`          | Inspect or retry jobs in the Dead Letter Queue.                             |
| `config`       | View or modify global configuration.                                        |

**Example Usage:**

```bash
queuectl enqueue '{"command":"echo Hello"}' --priority 10
queuectl worker start --count 2 --timeout 30s
queuectl list --state completed --show-output
queuectl dlq --retry job-123
```

---

### 5. **Job Lifecycle**

```
[PENDING] → [PROCESSING] → [COMPLETED]
                     ↓
                 (failure)
                     ↓
               [FAILED] → (retry)
                     ↓
            [DEAD] (after max_retries)
```

---

### 6. **Reliability & Safety Features**

| Feature                | Description                                     |
| ---------------------- | ----------------------------------------------- |
| **Persistence**        | All job states stored in SQLite (`queue.db`)    |
| **Retry & Backoff**    | Retries failed jobs with exponential delay      |
| **Dead Letter Queue**  | Permanent record of jobs that exhausted retries |
| **Timeout Handling**   | Jobs killed after `--timeout` duration          |
| **Graceful Shutdown**  | Ongoing jobs finish before exit on `Ctrl+C`     |
| **Priority Queues**    | Higher priority = earlier execution             |
| **Scheduled Jobs**     | `--delay` and `--run-at` supported              |
| **Job Output Logging** | Captured `stdout` and `stderr`                  |
| **Metrics / Stats**    | `queuectl stats` shows counts & averages        |

---

##  Data Flow

```
+-------------+         +------------------+         +---------------+
| CLI Command |  --->   | SQLite (JobRepo) |  --->   | Worker Process|
+-------------+         +------------------+         +---------------+
      ↑                         ↓                           ↓
 enqueue job             claim + execute job         complete or retry
      ↑                         ↓                           ↓
  show list/status  <--- update state <--- retry or DLQ transition
```

---

##  Design Decisions & Trade-offs

| Decision                          | Rationale                                                                    |
| --------------------------------- | ---------------------------------------------------------------------------- |
| **SQLite + GORM**                 | Simple, persistent, and transactional storage without external dependencies. |
| **Cobra for CLI**                 | Provides structured commands, flags, and help text for a polished UX.        |
| **Goroutines for Workers**        | Lightweight concurrency; easy to scale on a single machine.                  |
| **Context for Shutdowns**         | Ensures ongoing jobs finish gracefully on termination.                       |
| **Exponential Backoff**           | Prevents hot loops and allows recovery from transient failures.              |
| **Atomic Claims via Transaction** | Eliminates race conditions and duplicate job execution.                      |

---

##  Extensibility

This system is built with clear separations of concern and is easily extendable:

| Future Feature         | Possible Extension                            |
| ---------------------- | --------------------------------------------- |
| Web Dashboard          | Serve `/metrics` or Web UI over HTTP          |
| Distributed Processing | Replace SQLite with PostgreSQL or Redis       |
| Job Cancellation       | Extend worker to monitor cancellation context |
| Priority Tiers         | Add weighted job queueing                     |
| Prometheus Integration | Expose `JobMetrics()` as Prometheus metrics   |

---

##  Tech Stack

* **Language:** Go 1.23+
* **CLI Framework:** Cobra
* **ORM / DB:** GORM + SQLite
* **Concurrency:** Goroutines + Context
* **Persistence File:** `queue.db`
* **Testing:** Go `testing` package (`TestCreateAndFetchJob` etc.)

---

##  Summary

`queuectl` provides a **robust, maintainable, and feature-rich job queue system** with:

* Persistent storage
* Parallel workers
* Safe retry logic
* Graceful shutdown
* DLQ support
* Configurable priority and delay

