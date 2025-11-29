// package services

// import (
// 	"encoding/json"
// 	"fmt"
// 	"go-multi-tenant/config"
// 	"go-multi-tenant/models"
// 	"go-multi-tenant/queue"
// 	"go-multi-tenant/repositories"
// 	"log"
// 	"time"
// )

// type QueueService struct {
// 	userQueue   *queue.Queue
// 	workerCount int
// 	userService *UserService
// }

// func NewQueueService(userService *UserService, workerCount int) *QueueService {
// 	return &QueueService{
// 		userQueue:   queue.NewQueue(),
// 		workerCount: workerCount,
// 		userService: userService,
// 	}
// }

// type CreateUserJobData struct {
// 	TenantID    uint               `json:"tenant_id"`
// 	Request     *CreateUserRequest `json:"request"`
// 	CurrentUser *models.User       `json:"current_user"`
// }

// type CreateUserResponse struct {
// 	User  *models.User `json:"user,omitempty"`
// 	Error string       `json:"error,omitempty"`
// 	JobID string       `json:"job_id"`
// }

// func (qs *QueueService) StartWorkers() {
// 	for i := 0; i < qs.workerCount; i++ {
// 		go qs.worker(i + 1)
// 	}
// 	log.Printf("Started %d queue workers", qs.workerCount)
// }

// func (qs *QueueService) worker(workerID int) {
// 	for {
// 		job, ok := qs.userQueue.Dequeue()
// 		if !ok {
// 			continue
// 		}

// 		log.Printf("Worker %d processing job %s", workerID, job.ID)

// 		var jobData CreateUserJobData
// 		if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
// 			log.Printf("Worker %d: Error unmarshaling job data: %v", workerID, err)
// 			qs.userQueue.MarkCompleted(job.ID, "", err)
// 			continue
// 		}

// 		masterDB := config.GetMasterDB()
// 		tenantRepo := repositories.NewTenantRepository(masterDB)

// 		tenant, err := tenantRepo.GetByID(jobData.TenantID)
// 		if err != nil {
// 			log.Printf("Worker %d: Failed to find tenant: %v", workerID, err)
// 			qs.userQueue.MarkCompleted(job.ID, "", err)
// 			continue
// 		}

// 		tenantDB, err := config.TenantManager.GetTenantDB(tenant)
// 		if err != nil {
// 			log.Printf("Worker %d: Failed to connect to tenant DB: %v", workerID, err)
// 			qs.userQueue.MarkCompleted(job.ID, "", err)
// 			continue
// 		}

// 		user, err := qs.userService.CreateUser(
// 			tenantDB,
// 			jobData.TenantID,
// 			jobData.Request,
// 			jobData.CurrentUser,
// 		)

// 		// Store result
// 		var result string
// 		if err == nil && user != nil {
// 			result = fmt.Sprintf("User created successfully: %s (ID: %d)", user.Username, user.ID)
// 		}

// 		qs.userQueue.MarkCompleted(job.ID, result, err)

// 		if err != nil {
// 			log.Printf("Worker %d: User creation failed for job %s: %v", workerID, job.ID, err)
// 		} else {
// 			log.Printf("Worker %d: User created successfully for job %s: %s", workerID, job.ID, user.Username)
// 		}
// 	}
// }

// func (qs *QueueService) EnqueueUserCreation(tenantID uint, req *CreateUserRequest, currentUser *models.User) (string, error) {
// 	jobData := CreateUserJobData{
// 		TenantID:    tenantID,
// 		Request:     req,
// 		CurrentUser: currentUser,
// 	}

// 	jsonData, err := json.Marshal(jobData)
// 	if err != nil {
// 		return "", err
// 	}

// 	jobID := fmt.Sprintf("user_create_%d_%d", tenantID, time.Now().UnixNano())

// 	job := queue.Job{
// 		ID:        jobID,
// 		TenantID:  tenantID,
// 		Data:      string(jsonData),
// 		CreatedAt: time.Now(),
// 	}

// 	qs.userQueue.Enqueue(job)

// 	log.Printf("Job enqueued: %s, Queue size: %d", jobID, qs.userQueue.Size())

// 	return jobID, nil
// }

// func (qs *QueueService) GetQueueStats() map[string]interface{} {
// 	allJobs := qs.userQueue.GetAllJobs()

// 	return map[string]interface{}{
// 		"queue_size":     qs.userQueue.Size(),
// 		"completed_size": qs.userQueue.CompletedSize(),
// 		"is_empty":       qs.userQueue.IsEmpty(),
// 		"worker_count":   qs.workerCount,
// 		"pending_jobs":   len(allJobs["pending"]),
// 		"completed_jobs": len(allJobs["completed"]),
// 	}
// }

// func (qs *QueueService) GetJobStatus(jobID string) (queue.Job, bool) {
// 	return qs.userQueue.GetJobStatus(jobID)
// }

// func (qs *QueueService) GetAllJobs() map[string][]queue.Job {
// 	return qs.userQueue.GetAllJobs()
// }

//	func (qs *QueueService) CleanupCompletedJobs() int {
//		return qs.userQueue.CleanupCompletedJobs()
//	}
package services
