package routes

import (
	"encoding/json"
	l "pgnextgenconsumer/config"
	"pgnextgenconsumer/mappers"
	queue "pgnextgenconsumer/queue"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	addRequest = "/request"
	getRequest = "/:request_id"
)

type APILogRoute struct {
	Queue *queue.EventQueue
}

func ConfigureEventHandlerHTTP(e *echo.Group, q *queue.EventQueue) {
	handler := &APILogRoute{Queue: q}
	handler.AddHandlers(e)
}

func (_this *APILogRoute) AddHandlers(e *echo.Group) {
	e.POST(addRequest, _this.AddRequest)
}

func (_this *APILogRoute) AddRequest(c echo.Context) error {
	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		l.H.Error(c.Request().Context(), "Cannot decode json body", err)
		return c.JSON(400, err.Error())
	}
	event, err := mappers.Create(jsonBody)
	if err != nil {
		l.H.Error(c.Request().Context(), "Error while mapping json to Event", err)
		return c.JSON(400, err.Error())
	}
	taskBytes, err := json.Marshal(event)
	if err != nil {
		l.H.Error(c.Request().Context(), "Cannot serialize Event to bytes", err)
		return c.JSON(400, err.Error())
	}
	err = _this.Queue.TaskQueue.PublishBytes(taskBytes)
	if err != nil {
		l.H.Error(c.Request().Context(), "Failed in publishing to queue", err)
		return c.JSON(500, err.Error())
	}
	l.H.Info(c.Request().Context(), "Event successfully processed", zap.Any("event", event))
	return c.JSON(200, &event)
}
