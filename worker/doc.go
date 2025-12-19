// Package worker provides functionality for background task processing.
//
// It uses the asynq library, which is built on top of Redis, to manage
// asynchronous tasks. The package is divided into two main parts:
//
// - TaskDistributor: Responsible for creating and enqueuing tasks into Redis.
//
// - TaskProcessor: Responsible for picking up tasks from Redis and executing
// them. It supports multiple queues with different priorities (e.g., critical, default).
package worker
