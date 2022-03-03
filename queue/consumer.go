package queue

import (
	"context"
	"encoding/json"
	"fmt"
	l "pgnextgenconsumer/config"
	"pgnextgenconsumer/emitter"
	"pgnextgenconsumer/mappers"
	"time"

	"github.com/adjust/rmq/v4"
	"go.uber.org/zap"
)

type Consumer struct {
	name            string
	count           int
	before          time.Time
	Emitter         *emitter.Database
	reportBatchSize int
}

func NewConsumer(tag int, config l.ConsumerConfig, emitter *emitter.Database) *Consumer {
	return &Consumer{
		name:            fmt.Sprintf("consumer%d", tag),
		count:           0,
		before:          time.Now(),
		Emitter:         emitter,
		reportBatchSize: config.ReportBatchSize,
	}
}

func InitConsumer(c context.Context, config l.ConsumerConfig, q *EventQueue, e *emitter.Database) {
	var queue = q.TaskQueue
	var pollDuration = time.Duration(config.PollDuration) * time.Millisecond
	if err := queue.StartConsuming(config.PrefetchLimit, pollDuration); err != nil {
		panic(err)
	}

	for i := 0; i < config.NumConsumers; i++ {
		name := fmt.Sprintf("{consumer} %d", i)
		if _, err := queue.AddBatchConsumer(name, config.PrefetchLimit, pollDuration,
			NewConsumer(i, config, e)); err != nil {
			l.H.Error(c, "Error encountered when adding consumer", err)
			panic(err)
		}
	}
	l.H.Info(c, "Consumers initiated successfully: ", zap.Int("num", config.NumConsumers))
	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, syscall.SIGINT)
	// defer signal.Stop(signals)

	// <-signals // wait for signal
	// go func() {
	// 	<-signals // hard exit on second signal (in case shutdown gets stuck)
	// 	os.Exit(1)
	// }()

	// <-connection.StopAllConsuming() // wait for all Consume() calls to finish
}

func (consumer *Consumer) Consume(deliveries rmq.Deliveries) {
	fmt.Println("in consume")
	ctx := context.Background()
	var events = make([]mappers.Event, 0)
	for _, delivery := range deliveries {
		payload := delivery.Payload()
		event := &mappers.Event{}
		err := json.Unmarshal([]byte(payload), event)
		if err != nil {
			l.H.Error(ctx, "failure in creating event in consumer, ", err)
			continue
		}
		events = append(events, *event)
	}
	fmt.Println("looping done", events)

	l.H.Info(ctx, "Consuming event...", zap.Int("eventLen", len(events)))
	err := consumer.Emitter.BatchEmit(events)
	if err != nil {
		l.H.Error(ctx, "Failed to emit event...", err)
		return
	}
	for _, delivery := range deliveries {
		delivery.Ack()
	}
	var newIndex = consumer.count + len(deliveries)/consumer.reportBatchSize
	var prevIndex = consumer.count / consumer.reportBatchSize
	if newIndex-prevIndex == 1 {
		// Don't want the metrics as of now
		//duration := time.Since(consumer.before)
		//consumer.before = time.Now()
		//perSecond := time.Second / (duration / reportBatchSize)
		l.H.Info(ctx, "Consumer statistics: ", zap.String("name", consumer.name),
			zap.Int("count", consumer.count))
	}
	consumer.count += len(deliveries)
}
