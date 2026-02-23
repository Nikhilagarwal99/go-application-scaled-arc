package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"go.uber.org/zap"
)

// Task type constants — unique string identifier for each task type.
// Think of these like queue names in BullMQ.
const (
	TypeVerifyEmail  = "email:verify"
	TypeWelcomeEmail = "email:welcome"
)

// ---- Payloads ---------------------------------------------------------------
// Each task has its own payload struct — strongly typed, no raw maps.

type VerifyEmailPayload struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type WelcomeEmailPayload struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ---- Processor --------------------------------------------------------------
// EmailProcessor holds dependencies needed to process email tasks.
// It's the equivalent of a BullMQ worker process function.

type EmailProcessor struct {
	mailService *utils.MailService
}

func NewEmailProcessor(mailService *utils.MailService) *EmailProcessor {
	return &EmailProcessor{mailService: mailService}
}

// ProcessVerifyEmail handles the email:verify task.
// Asynq calls this function when a verify email job is dequeued.
// If this returns an error — asynq retries automatically.
func (p *EmailProcessor) ProcessVerifyEmail(ctx context.Context, t *asynq.Task) error {
	var payload VerifyEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		// Don't retry on bad payload — it will never succeed
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	logger.Info("processing verify email task",
		zap.String("email", payload.Email),
	)

	if err := p.mailService.SendVerifyEmail(payload.Email, payload.OTP); err != nil {
		// Return error — asynq will retry based on retry policy
		logger.Warn("verify email task failed — will retry",
			zap.String("email", payload.Email),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	logger.Info("verify email sent successfully",
		zap.String("email", payload.Email),
	)
	return nil
}

// ProcessWelcomeEmail handles the email:welcome task.
func (p *EmailProcessor) ProcessWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	var payload WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	logger.Info("processing welcome email task",
		zap.String("email", payload.Email),
	)

	// TODO: implement welcome email in mailjet.go
	// p.mailService.SendWelcomeEmail(payload.Email, payload.Name)

	return nil
}
