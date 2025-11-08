package job

import (
	"time"

	"gorm.io/gorm"
)

type JobState string

const (
	StatePending    JobState = "pending"
	StateProcessing JobState = "processing"
	StateCompleted  JobState = "completed"
	StateFailed     JobState = "failed"
	StateDead       JobState = "dead"
)

type Job struct {
	ID         string         `json:"id" gorm:"primaryKey;size:64"`
	Command    string         `json:"command" gorm:"not null"`
	State      JobState       `json:"state" gorm:"index;not null;default:'pending'"`
	Attempts   int32          `json:"attempts" gorm:"not null;default:0"`
	MaxRetries int32          `json:"max_retries" gorm:"not null;default:3"`
	Output     string         `json:"output"`
	Duration   float64        `json:"duration"`
	Priority int `json:"priority" gorm:"default:0;index"` 
	RunAt      *time.Time     `json:"run_at,omitempty"`
	LastError  *string        `json:"last_error,omitempty"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
