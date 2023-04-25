package main

import (
	"fmt"
	"github.com/kneu-messenger-pigeon/events"
	"io"
)

type ScoreChangedEventHandler struct {
	out              io.Writer
	serviceContainer *ServiceContainer
}

func (handler *ScoreChangedEventHandler) GetExpectedMessageKey() string {
	return events.ScoreChangedEventName
}

func (handler *ScoreChangedEventHandler) GetExpectedEventType() any {
	return &events.ScoreChangedEvent{}
}

func (handler *ScoreChangedEventHandler) Commit() error {
	return nil
}

func (handler *ScoreChangedEventHandler) Handle(s any) error {
	event := s.(*events.ScoreChangedEvent)
	if handler.serviceContainer != nil && handler.serviceContainer.ClientController != nil {
		go handler.callControllerAction(event)
	}

	return nil
}

func (handler *ScoreChangedEventHandler) callControllerAction(event *events.ScoreChangedEvent) {
	err := handler.serviceContainer.ClientController.ScoreChangedAction(event)
	if err != nil {
		_, _ = fmt.Fprintf(handler.out, "ScoreChangedAction return error: %v", err)
	}
}
