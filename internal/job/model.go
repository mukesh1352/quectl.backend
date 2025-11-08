package job
import (
	"time"
	"gorm.io/gorm"
)

type JobState string

const (
	StatePending JobState="pending"
	StateProcessing JobState="processing"
	StateCompleted JobState="completed"
	StateFailed JobState="Failed"
	StateDead JobState="Dead"
)

type Job struct{
	ID string `json:"id" gorm:"primaryKey;size:64`
	Command    string         `json:"command" gorm:"not null"` // shell command
	State      JobState       `json:"state" gorm:"index;not null;default:'pending'"`
	Attempts   int32          `json:"attempts" gorm:"not null;default:0"`
	MaxRetries int32          `json:"max_retries" gorm:"not null;default:3"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	RunAt      *time.Time     `json:"run_at,omitempty"` // next execution time for retry
	LastError  *string        `json:"last_error,omitempty"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
