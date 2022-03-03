package mappers

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type ExternalEvent struct {
	MerchantId int64
	URI        string
	ReqType    string
	Headers    map[string]interface{}
	Request    map[string]interface{}
	Response   map[string]interface{}
	Source     string
	AddedOn    string
}

type Event struct {
	MerchantId int64
	URI        string
	ReqType    string
	Headers    string
	Request    string
	Response   string
	Source     string
	AddedOn    time.Time
}

type EventMapper interface {
	Create() (*Event, error)
}

func Create(data map[string]interface{}) (*Event, error) {
	e := &ExternalEvent{}
	transcode(data, &e)
	event, err := validateAndConvert(*e)
	if err != nil {
		return event, err
	}
	return event, nil
}

func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(in)
	json.NewDecoder(buf).Decode(out)
}

func validateAndConvert(event ExternalEvent) (*Event, error) {
	if event.MerchantId <= 0 {
		return nil, errors.New("event does not have merchant id")
	}
	if event.URI == "" {
		return nil, errors.New("event does not have uri")
	}
	if event.ReqType == "" {
		return nil, errors.New("event does not have request type")
	}
	if event.Response == nil {
		return nil, errors.New("event has no resposne")
	}
	if event.Source == "" {
		return nil, errors.New("Event has no source")
	}
	requestString, err := json.Marshal(event.Request)
	if err != nil {
		return nil, errors.New("request map cannot be converted to string")
	}
	responseString, err := json.Marshal(event.Response)
	if err != nil {
		return nil, errors.New("response map cannot be converted to string")
	}
	headerString, err := json.Marshal(event.Headers)
	if err != nil {
		return nil, errors.New("headers map cannot be converted to string")
	}
	t1, err := time.Parse("2006-01-02 15:04:05", event.AddedOn)
	if err != nil {
		return nil, errors.New("addedOn cannot be converted to Datetime" + event.AddedOn)
	}

	return &Event{
		MerchantId: event.MerchantId,
		URI:        event.URI,
		ReqType:    event.ReqType,
		Headers:    string(headerString),
		Request:    string(requestString),
		Response:   string(responseString),
		Source:     event.Source,
		AddedOn:    t1,
	}, nil
}
