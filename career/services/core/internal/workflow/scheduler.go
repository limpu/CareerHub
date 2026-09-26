package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScheduleJob represents a scheduled recurring or one-off task.
type ScheduleJob struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspace_id"`
	TaskType       string          `json:"task_type"`
	Payload        json.RawMessage `json:"payload"`
	Interval       time.Duration   `json:"interval"`
	NextRunAt      time.Time       `json:"next_run_at"`
	LastRunAt      *time.Time      `json:"last_run_at,omitempty"`
	LeaseHolder    string          `json:"lease_holder,omitempty"`
	LeaseExpiresAt *time.Time      `json:"lease_expires_at,omitempty"`
	Enabled        bool            `json:"enabled"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// SchedulerRepository defines persistent operations for scheduled jobs and leases.
type SchedulerRepository interface {
	SaveJob(ctx context.Context, job *ScheduleJob) error
	GetJob(ctx context.Context, id string) (*ScheduleJob, error)
	GetDueJobs(ctx context.Context, now time.Time, limit int) ([]*ScheduleJob, error)
	AcquireLease(ctx context.Context, jobID, workerID string, now time.Time, duration time.Duration) (*ScheduleJob, error)
	HeartbeatLease(ctx context.Context, jobID, workerID string, now time.Time, extendDuration time.Duration) error
	ReleaseLease(ctx context.Context, jobID, workerID string, nextRunAt time.Time) error
}

// MemorySchedulerRepository provides an in-memory, thread-safe implementation of SchedulerRepository.
type MemorySchedulerRepository struct {
	mu   sync.RWMutex
	jobs map[string]*ScheduleJob
}

func NewMemorySchedulerRepository() *MemorySchedulerRepository {
	return &MemorySchedulerRepository{
		jobs: make(map[string]*ScheduleJob),
	}
}

func (r *MemorySchedulerRepository) SaveJob(ctx context.Context, job *ScheduleJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now

	// Clone to avoid concurrent mutation issues
	clone := *job
	r.jobs[job.ID] = &clone
	return nil
}

func (r *MemorySchedulerRepository) GetJob(ctx context.Context, id string) (*ScheduleJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	clone := *job
	return &clone, nil
}

func (r *MemorySchedulerRepository) GetDueJobs(ctx context.Context, now time.Time, limit int) ([]*ScheduleJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var due []*ScheduleJob
	for _, job := range r.jobs {
		if !job.Enabled {
			continue
		}
		// A job is due if next_run_at <= now AND (no lease or lease expired)
		isDue := !job.NextRunAt.After(now)
		leaseFree := job.LeaseHolder == "" || (job.LeaseExpiresAt != nil && !job.LeaseExpiresAt.After(now))

		if isDue && leaseFree {
			clone := *job
			due = append(due, &clone)
			if len(due) >= limit {
				break
			}
		}
	}
	return due, nil
}

func (r *MemorySchedulerRepository) AcquireLease(ctx context.Context, jobID, workerID string, now time.Time, duration time.Duration) (*ScheduleJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}

	if !job.Enabled {
		return nil, fmt.Errorf("job %s is disabled", jobID)
	}

	// Check if lease is currently valid and held by another worker
	if job.LeaseHolder != "" && job.LeaseExpiresAt != nil && job.LeaseExpiresAt.After(now) && job.LeaseHolder != workerID {
		return nil, ErrLeaseExpiredOrHeld
	}

	expiry := now.Add(duration)
	job.LeaseHolder = workerID
	job.LeaseExpiresAt = &expiry
	job.UpdatedAt = now

	clone := *job
	return &clone, nil
}

func (r *MemorySchedulerRepository) HeartbeatLease(ctx context.Context, jobID, workerID string, now time.Time, extendDuration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return ErrJobNotFound
	}

	if job.LeaseHolder != workerID {
		return ErrInvalidLease
	}

	if job.LeaseExpiresAt != nil && !job.LeaseExpiresAt.After(now) {
		return ErrLeaseExpiredOrHeld
	}

	expiry := now.Add(extendDuration)
	job.LeaseExpiresAt = &expiry
	job.UpdatedAt = now
	return nil
}

func (r *MemorySchedulerRepository) ReleaseLease(ctx context.Context, jobID, workerID string, nextRunAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return ErrJobNotFound
	}

	if job.LeaseHolder != workerID {
		return ErrInvalidLease
	}

	now := time.Now().UTC()
	job.LeaseHolder = ""
	job.LeaseExpiresAt = nil
	job.LastRunAt = &now
	job.NextRunAt = nextRunAt
	job.UpdatedAt = now
	return nil
}

// SchedulerService orchestrates job scheduling, exclusive leases and heartbeats (REQ-019, AT-018).
type SchedulerService struct {
	repo SchedulerRepository
}

func NewSchedulerService(repo SchedulerRepository) *SchedulerService {
	return &SchedulerService{repo: repo}
}

// RegisterSchedule creates or updates a scheduled job.
func (s *SchedulerService) RegisterSchedule(ctx context.Context, job *ScheduleJob) (*ScheduleJob, error) {
	if job.WorkspaceID == "" || job.TaskType == "" {
		return nil, fmt.Errorf("workspace_id and task_type are required")
	}
	if job.NextRunAt.IsZero() {
		job.NextRunAt = time.Now().UTC()
	}
	job.Enabled = true
	if err := s.repo.SaveJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

// AcquireDueJobs finds and leases due jobs for a worker (AT-018).
func (s *SchedulerService) AcquireDueJobs(ctx context.Context, workerID string, now time.Time, duration time.Duration, limit int) ([]*ScheduleJob, error) {
	due, err := s.repo.GetDueJobs(ctx, now, limit)
	if err != nil {
		return nil, err
	}

	var acquired []*ScheduleJob
	for _, job := range due {
		leased, err := s.repo.AcquireLease(ctx, job.ID, workerID, now, duration)
		if err == nil && leased != nil {
			acquired = append(acquired, leased)
		}
	}
	return acquired, nil
}

// HeartbeatLease renews a worker's active lease.
func (s *SchedulerService) HeartbeatLease(ctx context.Context, jobID, workerID string, extendDuration time.Duration) error {
	return s.repo.HeartbeatLease(ctx, jobID, workerID, time.Now().UTC(), extendDuration)
}

// ReleaseLease finishes job processing and computes the next due time.
func (s *SchedulerService) ReleaseLease(ctx context.Context, jobID, workerID string, nextInterval time.Duration) error {
	now := time.Now().UTC()
	nextRunAt := now.Add(nextInterval)
	return s.repo.ReleaseLease(ctx, jobID, workerID, nextRunAt)
}

// GetJob fetches a schedule by ID.
func (s *SchedulerService) GetJob(ctx context.Context, id string) (*ScheduleJob, error) {
	return s.repo.GetJob(ctx, id)
}
