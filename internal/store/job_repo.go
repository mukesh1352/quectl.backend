package store

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"queuectl.backend/internal/job"
)

type JobRepo struct {
	db *gorm.DB
}

// NewJobRepo creates a new repository instance
func NewJobRepo(db *gorm.DB) *JobRepo {
	return &JobRepo{db: db}
}

// Create inserts a new job into the database
func (r *JobRepo) Create(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	return r.db.Create(j).Error
}

// FindPending fetches the next pending job (FIFO + priority-aware)
func (r *JobRepo) FindPending() (*job.Job, error) {
	var j job.Job
	tx := r.db.
		Where("state = ? AND (run_at IS NULL OR run_at <= ?)", job.StatePending, time.Now().UTC()).
		Order("priority DESC, created_at ASC"). // ✅ Priority-first ordering
		First(&j)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &j, nil
}

// Update updates an existing job or its metadata
func (r *JobRepo) Update(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.UpdatedAt = time.Now().UTC()
	return r.db.Save(j).Error
}

// Processing marks a job as being processed by a worker
func (r *JobRepo) Processing(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateProcessing
	return r.Update(j)
}

// MarkCompleted marks a job as successfully done
func (r *JobRepo) MarkCompleted(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateCompleted
	return r.Update(j)
}

// Failed handles retry logic or moves job to DLQ if max retries reached
func (r *JobRepo) Failed(j *job.Job, errMsg string, baseDelay time.Duration) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.Attempts++
	j.LastError = &errMsg
	now := time.Now().UTC()
	if j.Attempts >= j.MaxRetries {
		j.State = job.StateDead
		j.RunAt = nil
	} else {
		j.State = job.StateFailed
		delay := baseDelay * time.Duration(1<<(j.Attempts-1)) // exponential backoff
		nextRun := now.Add(delay)
		j.RunAt = &nextRun
	}
	j.UpdatedAt = now
	return r.db.Save(j).Error
}

// ListJobs retrieves jobs filtered by state and priority
func (r *JobRepo) ListJobs(states []job.JobState, limit, offset int32, newestFirst bool) ([]job.Job, error) {
	var jobs []job.Job
	if limit <= 0 {
		limit = 100
	}
	order := "priority DESC, created_at ASC" // ✅ always priority-first
	if newestFirst {
		order = "priority DESC, created_at DESC"
	}
	query := r.db.Order(order).Offset(int(offset))
	if len(states) > 0 {
		query = query.Where("state IN ?", states)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

// PreventRaceCondition ensures only one worker claims a pending job
func (r *JobRepo) PreventRaceCondition(workerId string) (*job.Job, error) {
	now := time.Now().UTC()
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var j job.Job
	err := tx.
		Where("state = ? AND (run_at IS NULL OR run_at <= ?)", job.StatePending, now).
		Order("priority DESC, created_at ASC").
		Limit(1).
		Take(&j).Error

	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	res := tx.Model(&job.Job{}).
		Where("id = ? AND state = ?", j.ID, job.StatePending).
		Updates(map[string]interface{}{
			"state":      job.StateProcessing,
			"updated_at": now,
		})

	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, nil
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	j.State = job.StateProcessing
	return &j, nil
}

func (r *JobRepo) DB() *gorm.DB {
	return r.db
}

type MetricsSummary struct {
	Total       int64
	Pending     int64
	Processing  int64
	Completed   int64
	Failed      int64
	Dead        int64
	AvgDuration float64
	AvgRetries  float64
}

// JobMetrics returns overall queue statistics and averages.
func (r *JobRepo) JobMetrics() (MetricsSummary, error) {
	var summary MetricsSummary

	// Count by state
	if err := r.db.Model(&job.Job{}).Count(&summary.Total).Error; err != nil {
		return summary, err
	}
	r.db.Model(&job.Job{}).Where("state = ?", job.StatePending).Count(&summary.Pending)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateProcessing).Count(&summary.Processing)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateCompleted).Count(&summary.Completed)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateFailed).Count(&summary.Failed)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateDead).Count(&summary.Dead)

	// Calculate averages (avoid NULL issues)
	r.db.Model(&job.Job{}).Select("COALESCE(AVG(duration), 0)").Scan(&summary.AvgDuration)
	r.db.Model(&job.Job{}).Select("COALESCE(AVG(attempts), 0)").Scan(&summary.AvgRetries)

	return summary, nil
}
