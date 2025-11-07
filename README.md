# Queuectl Backend
Here we are creating this queuctl project for job worker management

- **To run the go project**
```zsh
go run .
```
- **To build the go project**
```zsh
go build
```
 the binary file is created.
- **To run the Binary file**
```zsh
./binary_file
```
### Why Golang
- Golang is one of the fastest working programming languages
- Golang is mainly used for micro-services making it more useful for jobs scheduling and all

## **Job Structure**
```JSON
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
#  Job Lifecycle & CLI Reference

##  Job Lifecycle

| **State** | **Description** |
|------------|-----------------|
| `pending` | Waiting to be picked up by a worker |
| `processing` | Currently being executed |
| `completed` | Successfully executed |
| `failed` | Failed, but retryable |
| `dead` | Permanently failed (moved to DLQ) |

---

##  CLI Commands Reference

| **Category** | **Command Example** | **Description** |
|---------------|----------------------|------------------|
| **Enqueue** | `queuectl enqueue '{"id":"job1","command":"sleep 2"}'` | Add a new job to the queue |
| **Workers** | `queuectl worker start --count 3` | Start one or more workers |
|  | `queuectl worker stop` | Stop running workers gracefully |
| **Status** | `queuectl status` | Show summary of all job states & active workers |
| **List Jobs** | `queuectl list --state pending` | List jobs by state |
| **DLQ** | `queuectl dlq list` / `queuectl dlq retry job1` | View or retry DLQ jobs |
| **Config** | `queuectl config set max-retries 3` | Manage configuration (retry count, backoff, etc.) |

---
