package job

import "time"

type JobState string

const (
	StatePending    JobState = "pending"
	StateProcessing JobState = "processing"
	StateCompleted  JobState = "completed"
	StateFailed     JobState = "failed"
	StateDead       JobState = "dead"
)

type Job struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Command    string    `json:"command"`
	State      JobState  `json:"state"`
	Attempts   int32     `json:"attempts"`
	MaxRetries int32     `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
