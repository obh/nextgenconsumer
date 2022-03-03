package queue

import (
	"context"
	"os"
	"os/signal"
	l "pgnextgenconsumer/config"
	"syscall"

	"github.com/adjust/rmq/v4"
)

const (
	redisHost = "localhost:7000"
	queueName = "testqueue"
)

type EventQueue struct {
	TaskQueue rmq.Queue
}

func InitQueue(config l.RedisConfig, errChan chan<- error) (*EventQueue, error) {
	// errChan := make(chan error)
	redisHost := config.Hostname + ":" + config.Port
	connection, err := rmq.OpenConnection("my service", "tcp", redisHost, 1, errChan)
	ctx := context.Background()
	// r := redis.NewClient(&redis.Options{
	// 	Addr: redisHost,
	// })
	// connection, err := rmq.OpenConnectionWithRedisClient("my service", r, errChan)
	if err != nil {
		l.H.Error(ctx, "Failed in connecting to redis, ", err)
		return nil, err
	}
	taskQueue, err := connection.OpenQueue(queueName)
	if err != nil {
		return nil, err
	}
	return &EventQueue{TaskQueue: taskQueue}, nil
}

func WaitSignal() os.Signal {
	ch := make(chan os.Signal, 2)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
			return sig
		}
	}
}
