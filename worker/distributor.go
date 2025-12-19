package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeSendVerifyEmailTask(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor creates and returns a new TaskDistributor backed by a Redis client with the provided options.
func NewRedisTaskDistributor(options asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(options)
	return &RedisTaskDistributor{client: client}
}
