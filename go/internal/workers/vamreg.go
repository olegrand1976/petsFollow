package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/redis/go-redis/v9"
)

const (
	TaskVamregDeclare = "pharmacy:vamreg:declare"
	QueueCritical     = "critical"
)

// VamregTaskPayload is the Asynq payload for VAMReg declare.
type VamregTaskPayload struct {
	PracticeID string `json:"practiceId"`
	DAFID      string `json:"dafId"`
}

// VamregEnqueue pushes a declare task (or no-ops if client nil).
type VamregEnqueue struct {
	Client *asynq.Client
}

// EnqueueDeclare schedules the critical-queue VAMReg task.
func (e *VamregEnqueue) EnqueueDeclare(ctx context.Context, practiceID, dafID string) error {
	if e == nil || e.Client == nil {
		return fmt.Errorf("asynq_client_unavailable")
	}
	body, err := json.Marshal(VamregTaskPayload{PracticeID: practiceID, DAFID: dafID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TaskVamregDeclare, body,
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(10),
		asynq.TaskID("vamreg:"+dafID),
	)
	_, err = e.Client.EnqueueContext(ctx, task)
	if err != nil {
		// Duplicate TaskID while prior task is pending/active — treat as already in-flight.
		if strings.Contains(err.Error(), "task ID conflicts") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil
		}
	}
	return err
}

// NewAsynqClient builds a Redis-backed Asynq client after a Ping fail-fast.
// Isolation between staging/prod is via Redis DB in REDIS_ADDR (not a key prefix).
func NewAsynqClient(redisAddr string) (*asynq.Client, error) {
	redisAddr = strings.TrimSpace(redisAddr)
	if redisAddr == "" {
		return nil, fmt.Errorf("redis_addr_empty")
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() { _ = rdb.Close() }()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping %s: %w", redisAddr, err)
	}
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}), nil
}

// NewAsynqServer builds a worker server for pharmacy queues.
func NewAsynqServer(redisAddr string) *asynq.Server {
	redisAddr = strings.TrimSpace(redisAddr)
	if redisAddr == "" {
		return nil
	}
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 4,
			Queues: map[string]int{
				QueueCritical: 6,
				"default":     3,
			},
		},
	)
}

// RegisterVamregHandler wires ProcessDeclare onto the mux.
func RegisterVamregHandler(mux *asynq.ServeMux, decl *pharmacy.VamregDeclarer) {
	mux.HandleFunc(TaskVamregDeclare, func(ctx context.Context, t *asynq.Task) error {
		var p VamregTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
		if err := decl.ProcessDeclare(ctx, p.PracticeID, p.DAFID); err != nil {
			log.Printf("vamreg declare failed daf=%s: %v", p.DAFID, err)
			return err
		}
		return nil
	})
}
