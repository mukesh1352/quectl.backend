package store

import (
	"errors"
	"time"

	"github.com/mukesh1352/go-api-starter/internal/job"
	"gorm.io/gorm"
)

type JobRepo struct {
	db *gorm.DB
}

// job repository instance
func NewJobRepo(db *gorm.DB) *JobRepo {
	return &JobRepo{db: db}
}

// inserts new job instances into the database
func (r *JobRepo) Create(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	return r.db.Create(j).Error
}

// Fetches the next pending task
func (r *JobRepo) FindPending() (*job.Job, error) {
	var j job.Job
	tx := r.db.
		Where("state= ? AND (run_at IS NULL OR run_at<= ?)", job.StatePending, time.Now().UTC()).
		Order("created_at ASC").First(&j)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &j, nil
}

// Update updates an existing job or the metadata
func (r *JobRepo) Update(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.UpdatedAt = time.Now().UTC()
	return r.db.Save(j).Error
}

// Jobs Being Processed
func (r *JobRepo) Processing(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateProcessing
	return r.Update(j) //update the current processing job
}

// mark as completed
func (r *JobRepo) MarkCompleted(j *job.Job) error {
	if j == nil {
		return errors.New("job cannot be nil")
	}
	j.State = job.StateCompleted
	return r.Update(j)
}

// Failed Jobs. -  moved to the dead letter queue after the max retries
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
		delay := baseDelay * time.Duration(1<<(j.Attempts-1))
		nextRun := now.Add(delay)
		j.RunAt = &nextRun
	}
	j.UpdatedAt = now
	return r.db.Save(j).Error
}

// JobList - Listing the jobs which are present
func (r *JobRepo) ListJobs(states []job.JobState, limit, offset int32, newestFirst bool) ([]job.Job, error) {
	var jobs []job.Job
	if limit <= 0 {
		limit = 100
	}
	order := "created_at ASC"
	if newestFirst {
		order = "created_at DESC"
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

// this function maintains that the jobs are working and there are no race conditions
func (r *JobRepo) PreventRaceCondition(workerId string) (*job.Job, error) {
	now := time.Now().UTC()
	tx := r.db.Begin() // starting the transaction
	if tx.Error != nil {
		return nil, tx.Error
	}
	var j job.Job
	err := tx.
		Where("state=? AND (run_at IS NULL OR run_at<=?)", job.StatePending, now).
		Order("created_at ASC").
		Limit(1).
		Take(&j).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, nil
	}
	//updation automatic
	res := tx.Model(&job.Job{}).Where("id=? AND state=?", j.ID, job.StatePending).
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
	//commiting the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	j.State = job.StateProcessing
	return &j, nil
}
