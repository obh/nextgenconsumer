package main

import (

	//_con "pgnextgenconsumer/eventconsumer"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	config "pgnextgenconsumer/config"
	_db "pgnextgenconsumer/emitter"
	queue "pgnextgenconsumer/queue"
	event "pgnextgenconsumer/routes"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/labstack/echo/v4"
)

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

	queue.InitConsumer(context.Background(), cfg.ConsumerConfig, q, db)
	eventRoute := e.Group("v1")
	event.ConfigureEventHandlerHTTP(eventRoute, q)

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	logErrors(errChan)
	graceFullShutdown(e, errChan)
}

func graceFullShutdown(e *echo.Echo, errChan chan error) {
	fmt.Println("in greaceful shutdown")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	close(errChan)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
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

func logErrors(errChan chan error) {
	fmt.Println("in log errors")
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			fmt.Println(err.RedisErr.Error())
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}
