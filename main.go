package main

import (

	//_con "pgnextgenconsumer/eventconsumer"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	config "pgnextgenconsumer/config"
	_db "pgnextgenconsumer/emitter"
	queue "pgnextgenconsumer/queue"
	event "pgnextgenconsumer/routes"
	"syscall"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/labstack/echo/v4"
)

const (
	logHeader = `{"time":"${time_rfc3339_nano}","level":"${level}","prefix":"prefixs","file":"${short_file}","line":"${line}", "application": "pgnextgenconsumer"}`
)

var counter int32

func main() {

	cfg := config.LoadConfig()
	fmt.Println(cfg)

	err := config.Init(cfg.LogConfig)

	db := _db.InitDB(cfg.MySqlConfig)
	fmt.Println(db)

	errChan := make(chan error)
	q, err := queue.InitQueue(cfg.RedisConfig, errChan)
	if err != nil {
		config.H.Error(context.Background(), "Failed to connect to queue, killing myself", err)
		return
	}

	e := echo.New()
	e.Logger.SetHeader(logHeader)

	queue.InitConsumer(context.Background(), cfg.ConsumerConfig, q, db)
	eventRoute := e.Group("v1")
	event.ConfigureEventHandlerHTTP(eventRoute, q)

	//go e.Start(cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
	logErrors(errChan)
	//graceFullShutdown(e)
}

func graceFullShutdown(e *echo.Echo) {
	fmt.Println("in greaceful shutdown")
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	<-channel
	log.Println("Service has been shut down")
	ctx, done := context.WithTimeout(context.Background(), time.Duration(5)*time.Minute)
	defer done()
	fmt.Println("Go routines cleared successfully to shut down")
	err := e.Shutdown(ctx)
	if err != nil {
		os.Exit(5000)
	}
}

func redisConnShutdown(e *echo.Echo, errChan <-chan error) {
	ctx, done := context.WithTimeout(context.Background(), time.Duration(1)*time.Minute)
	defer done()
	err := e.Shutdown(ctx)
	if err != nil {
		os.Exit(5000)
	}
}

func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}
