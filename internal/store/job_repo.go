package store

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"queuectl.backend/internal/job"
)

// JobRepo handles all DB operations for jobs.
type JobRepo struct {
	db *gorm.DB
}

// NewJobRepo creates a new repository instance.
func NewJobRepo(db *gorm.DB) *JobRepo {
	return &JobRepo{db: db}
}

// Create inserts a new job into the database.
func (r *JobRepo) Create(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	return r.db.Create(j).Error
}

// FindPending fetches the next job ready for execution.
// It supports priority and run_at (delayed jobs).
func (r *JobRepo) FindPending() (*job.Job, error) {
	var j job.Job
	tx := r.db.
		Where("(state = ? OR state = ?) AND (run_at IS NULL OR run_at <= ?)",
			job.StatePending, job.StateFailed, time.Now().UTC()).
		Order("priority DESC, created_at ASC").
		First(&j)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &j, nil
}

// Update saves job updates (metadata, timestamps, etc.).
func (r *JobRepo) Update(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.UpdatedAt = time.Now().UTC()
	return r.db.Save(j).Error
}

// Processing marks a job as being processed by a worker.
func (r *JobRepo) Processing(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateProcessing
	return r.Update(j)
}

// MarkCompleted marks a job as successfully completed.
func (r *JobRepo) MarkCompleted(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateCompleted
	return r.Update(j)
}

// Failed handles retry or moves the job to the DLQ after max retries.
func (r *JobRepo) Failed(j *job.Job, errMsg string, baseDelay time.Duration) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}

	j.Attempts++
	j.LastError = &errMsg
	now := time.Now().UTC()

	if j.Attempts >= j.MaxRetries {
		// Move to DLQ
		j.State = job.StateDead
		j.RunAt = nil
	} else {
		// Schedule retry with exponential backoff
		j.State = job.StateFailed
		delay := baseDelay * time.Duration(1<<(j.Attempts-1))
		nextRun := now.Add(delay)
		j.RunAt = &nextRun
	}

	j.UpdatedAt = now
	return r.db.Save(j).Error
}

// ListJobs retrieves jobs filtered by states and sorted by priority + creation time.
func (r *JobRepo) ListJobs(states []job.JobState, limit, offset int32, newestFirst bool) ([]job.Job, error) {
	var jobs []job.Job

	if limit <= 0 {
		limit = 100
	}

	order := "priority DESC, created_at ASC"
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

// PreventRaceCondition ensures that only one worker safely claims a job.
// It includes retryable (failed) jobs once their run_at time is due.
func (r *JobRepo) PreventRaceCondition(workerId string) (*job.Job, error) {
	now := time.Now().UTC()
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var j job.Job
	err := tx.
		Where("(state = ? OR state = ?) AND (run_at IS NULL OR run_at <= ?)",
			job.StatePending, job.StateFailed, now).
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

	// ✅ allow claiming from both pending and failed states
	res := tx.Model(&job.Job{}).
		Where("id = ? AND (state = ? OR state = ?)",
			j.ID, job.StatePending, job.StateFailed).
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

// DB returns the underlying database instance.
func (r *JobRepo) DB() *gorm.DB {
	return r.db
}

// MetricsSummary aggregates queue metrics.
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

// JobMetrics returns system-wide metrics and averages.
func (r *JobRepo) JobMetrics() (MetricsSummary, error) {
	var summary MetricsSummary

	// Count states
	if err := r.db.Model(&job.Job{}).Count(&summary.Total).Error; err != nil {
		return summary, err
	}
	r.db.Model(&job.Job{}).Where("state = ?", job.StatePending).Count(&summary.Pending)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateProcessing).Count(&summary.Processing)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateCompleted).Count(&summary.Completed)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateFailed).Count(&summary.Failed)
	r.db.Model(&job.Job{}).Where("state = ?", job.StateDead).Count(&summary.Dead)

	// Calculate averages
	r.db.Model(&job.Job{}).Select("COALESCE(AVG(duration), 0)").Scan(&summary.AvgDuration)
	r.db.Model(&job.Job{}).Select("COALESCE(AVG(attempts), 0)").Scan(&summary.AvgRetries)

	return summary, nil
}
