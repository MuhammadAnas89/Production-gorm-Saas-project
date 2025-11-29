// package queue

// import (
// 	"sync"
// 	"time"
// )

// type JobStatus string

// const (
// 	JobStatusPending   JobStatus = "pending"
// 	JobStatusCompleted JobStatus = "completed"
// 	JobStatusFailed    JobStatus = "failed"
// )

// type Job struct {
// 	ID        string    `json:"id"`
// 	TenantID  uint      `json:"tenant_id"`
// 	Data      string    `json:"data"`
// 	Status    JobStatus `json:"status"`
// 	Result    string    `json:"result,omitempty"`
// 	Error     string    `json:"error,omitempty"`
// 	CreatedAt time.Time `json:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at"`
// }

// type Queue struct {
// 	jobs         []Job
// 	completed    []Job
// 	mu           sync.Mutex
// 	notEmpty     *sync.Cond
// 	maxCompleted int
// }

// func NewQueue() *Queue {
// 	q := &Queue{
// 		jobs:         make([]Job, 0),
// 		completed:    make([]Job, 0),
// 		maxCompleted: 1000,
// 	}
// 	q.notEmpty = sync.NewCond(&q.mu)

// 	go q.startCleanupWorker()

// 	return q
// }

// func (q *Queue) Enqueue(job Job) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	job.Status = JobStatusPending
// 	job.CreatedAt = time.Now()
// 	job.UpdatedAt = time.Now()

// 	q.jobs = append(q.jobs, job)
// 	q.notEmpty.Signal()
// }

// func (q *Queue) Dequeue() (Job, bool) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	for len(q.jobs) == 0 {
// 		q.notEmpty.Wait()
// 	}

// 	job := q.jobs[0]
// 	q.jobs = q.jobs[1:]
// 	return job, true
// }

// func (q *Queue) MarkCompleted(jobID string, result string, err error) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	for i, job := range q.jobs {
// 		if job.ID == jobID {

// 			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
// 			break
// 		}
// 	}

// 	completedJob := Job{
// 		ID:        jobID,
// 		Status:    JobStatusCompleted,
// 		Result:    result,
// 		UpdatedAt: time.Now(),
// 	}

// 	if err != nil {
// 		completedJob.Status = JobStatusFailed
// 		completedJob.Error = err.Error()
// 	}

// 	q.completed = append(q.completed, completedJob)

// 	if len(q.completed) > q.maxCompleted {
// 		q.completed = q.completed[len(q.completed)-q.maxCompleted:]
// 	}
// }

// func (q *Queue) GetJobStatus(jobID string) (Job, bool) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	for _, job := range q.jobs {
// 		if job.ID == jobID {
// 			return job, true
// 		}
// 	}

// 	for _, job := range q.completed {
// 		if job.ID == jobID {
// 			return job, true
// 		}
// 	}

// 	return Job{}, false
// }

// func (q *Queue) GetAllJobs() map[string][]Job {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	return map[string][]Job{
// 		"pending":   q.jobs,
// 		"completed": q.completed,
// 	}
// }

// func (q *Queue) Size() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.jobs)
// }

// func (q *Queue) CompletedSize() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.completed)
// }

// func (q *Queue) IsEmpty() bool {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.jobs) == 0
// }

// func (q *Queue) CleanupCompletedJobs() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	cutoffTime := time.Now().Add(-1 * time.Hour)
// 	initialCount := len(q.completed)

// 	filtered := make([]Job, 0)
// 	for _, job := range q.completed {
// 		if job.UpdatedAt.After(cutoffTime) {
// 			filtered = append(filtered, job)
// 		}
// 	}

// 	q.completed = filtered
// 	cleanedCount := initialCount - len(q.completed)

// 	return cleanedCount
// }

// func (q *Queue) startCleanupWorker() {
// 	ticker := time.NewTicker(30 * time.Minute)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		cleaned := q.CleanupCompletedJobs()
// 		if cleaned > 0 {

//				println("Queue cleanup: removed", cleaned, "completed jobs older than 1 hour")
//			}
//		}
//	}
package queue
