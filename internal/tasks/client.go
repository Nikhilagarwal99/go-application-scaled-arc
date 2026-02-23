package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client to enqueue tasks.
// Called from the service layer — services never import asynq directly,
// only this package. Keeps asynq isolated to the tasks package.
type Client struct {
	asynq *asynq.Client
}

func NewClient(redisAddr, redisPassword string, redisDB int) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	return &Client{asynq: client}
}

func (c *Client) Close() error {
	return c.asynq.Close()
}

// EnqueueVerifyEmail enqueues a verify email task.
//
// asynq.MaxRetry(3)       — retry up to 3 times on failure
// asynq.Queue("critical") — use the critical priority queue
func (c *Client) EnqueueVerifyEmail(email, otp string) error {
	payload, err := json.Marshal(VerifyEmailPayload{
		Email: email,
		OTP:   otp,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal verify email payload: %w", err)
	}

	task := asynq.NewTask(TypeVerifyEmail, payload,
		asynq.MaxRetry(3),
		asynq.Queue("critical"), // high priority — user is waiting for this
	)

	// Enqueue the task
	_, err = c.asynq.Enqueue(task)
	return err
}

// EnqueueWelcomeEmail enqueues a welcome email task.
//
// asynq.Queue("default") — lower priority than verify email
func (c *Client) EnqueueWelcomeEmail(email, name string) error {
	payload, err := json.Marshal(WelcomeEmailPayload{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal welcome email payload: %w", err)
	}

	task := asynq.NewTask(TypeWelcomeEmail, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
	)

	_, err = c.asynq.Enqueue(task)
	return err
}
