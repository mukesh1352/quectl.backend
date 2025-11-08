Absolutely — here’s the **clean, professional version of your `README.md`** with all your real screenshots integrated in the correct locations,
but with **no extra “Expected Output” text blocks or examples** — only the images (so it’s visually elegant and concise).

---

````markdown
# Queuectl Backend

**Queuectl** is a Go-based command-line background job queue system that manages job scheduling, worker processing, retries, and failure handling.
It provides a robust, persistent job queue with automatic retry mechanisms, exponential backoff, and a Dead Letter Queue (DLQ) for permanently failed jobs.

This project demonstrates a production-ready architecture for reliable background processing — built for simplicity, concurrency, and transparency.

---

## Table of Contents

1. [Overview](#overview)
2. [Setup Instructions](#setup-instructions)
3. [Usage Examples](#usage-examples)
4. [Architecture Overview](#architecture-overview)
5. [Job Lifecycle](#job-lifecycle)
6. [Assumptions and Trade-offs](#assumptions-and-trade-offs)
7. [Testing Instructions](#testing-instructions)
8. [Folder Structure](#folder-structure)
9. [Future Enhancements](#future-enhancements)
10. [License](#license)

---

## Overview

Queuectl provides a lightweight, persistent job queue with CLI-based management.

Key features include:
- Persistent storage with **SQLite**
- Multiple concurrent **workers**
- **Retry mechanism** with exponential backoff
- **Dead Letter Queue (DLQ)** for permanently failed jobs
- **Job scheduling/delays** using `--delay` or `--run-at`
- **Job priorities** (`--priority` flag)
- Configurable **timeouts** and retry bases
- Simple **web dashboard** for metrics and live monitoring

---

## Setup Instructions

### Prerequisites

- Go **v1.23+**
- SQLite3 installed locally
- Git for cloning the repository

### Clone and Build

```bash
git clone https://github.com/<your-username>/queuectl.backend.git
cd queuectl.backend
go mod tidy
go build -o queuectl
````

### Run the CLI

```bash
./queuectl --help
```

![CLI Help Output](output/test.png)

---

## Usage Examples

### 1. Enqueue a Job

```bash
go run main.go enqueue '{"command":"echo Hello World"}'
```

![Enqueue Job Success](output/enqueue_success.png)

---

### 2. Start Workers

```bash
go run main.go worker start --count 2 --timeout 30s
```

![Worker Success Logs](output/worker_success.png)

---

### 3. Failed Job and DLQ Example

```bash
go run main.go enqueue '{"command":"exit 1"}'
go run main.go worker start --timeout 5s --backoff-base 2s
go run main.go dlq
```

![DLQ Output](output/dlq.png)

---

### 4. View Job Stats

```bash
go run main.go stats
```

![Stats Summary](output/stats.png)

---

### 5. View All Jobs

```bash
go run main.go list
```

![List Completed Jobs](output/list_completed.png)

---

### 6. Database Verification

```bash
sqlite3 queue.db "select id, state, attempts, run_at from jobs;"
```

![SQLite Job Table Output](output/sql.png)

---

### 7. Web Dashboard

```bash
go run main.go web
```

Visit: [http://localhost:8080](http://localhost:8080)

![Web Dashboard](output/web.png)

---

## Architecture Overview

Queuectl follows a modular and layered architecture for clarity and scalability:

### 1. CLI Layer (`cmd/`)

Handles all commands, arguments, and user interactions using **Cobra**.

### 2. Storage Layer (`internal/store/`)

Implements persistent job storage in **SQLite** using **GORM ORM**.
Responsible for:

* Job lifecycle management
* Retry and DLQ transitions
* Priority and scheduling logic

### 3. Worker Layer (`internal/queue/`)

Manages concurrent job execution using goroutines:

* Executes system commands
* Handles retries via exponential backoff
* Implements timeouts
* Supports graceful shutdowns

---

### Job Lifecycle

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

### Job Structure

```json
{
  "id": "unique-job-id",
  "command": "echo 'Hello World'",
  "state": "pending",
  "attempts": 0,
  "max_retries": 3,
  "created_at": "2025-11-04T10:30:00Z",
  "updated_at": "2025-11-04T10:30:00Z"
}
```

---

## Assumptions and Trade-offs

* **SQLite** chosen for local persistence and simplicity
  (production setups should use PostgreSQL or Redis).
* CLI-first design ensures easy automation and debugging.
* Exponential backoff prevents retry overload.
* Worker coordination is currently **single-node**.
* Shell command execution assumes a trusted environment.
* Configurable retry count, delay, and timeout enhance flexibility.

---

## Testing Instructions

### Run Unit Tests

```bash
go test ./...
```

![Go Test Output](output/go_test.png)

---

### Manual Functional Test

```bash
# 1. Enqueue a job
go run main.go enqueue '{"command":"echo Hello"}'

# 2. Start worker
go run main.go worker start --timeout 5s --backoff-base 2s

# 3. Verify job completion
go run main.go list --state completed
```

![Output Screenshot](output/output.png)

---

## Folder Structure

```
.
├── DESIGN.md
├── LICENSE
├── README.md
├── cmd
│   ├── common.go
│   ├── config.go
│   ├── dlq.go
│   ├── enqueue.go
│   ├── list.go
│   ├── root.go
│   ├── stats.go
│   ├── status.go
│   ├── web.go
│   └── worker.go
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── model.go
│   ├── job
│   │   ├── job_test.go
│   │   ├── model.go
│   │   └── queue.db
│   ├── queue
│   │   ├── executor.go
│   │   └── worker.go
│   └── store
│       ├── job_repo.go
│       └── store.go
├── main.go
├── output
│   ├── dlq.png
│   ├── enqueue_success.png
│   ├── go_test.png
│   ├── list_completed.png
│   ├── output.png
│   ├── sql.png
│   ├── stats.png
│   ├── test.png
│   ├── web.png
│   └── worker_success.png
├── queue.db
└── scripts
    └── test_demo.sh
```

---

## Future Enhancements

* Distributed worker coordination
* REST API for remote management
* WebSocket live updates for dashboard
* Role-based access control
* Pause/resume job support
* Integration with message queues (RabbitMQ, Kafka)

---

## License

This project is licensed under the **MIT License**.
See the [LICENSE](LICENSE) file for details.

```

