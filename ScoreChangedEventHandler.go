package main

import (
	"github.com/kneu-messenger-pigeon/events"
)

type ScoreChangedEventHandler struct {
	eventQueue chan *events.ScoreChangedEvent
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
	if handler.eventQueue != nil {
		handler.eventQueue <- event
	}

	return nil
}

func (handler *ScoreChangedEventHandler) GetEventQueue() <-chan *events.ScoreChangedEvent {
	if handler.eventQueue == nil {
		handler.eventQueue = make(chan *events.ScoreChangedEvent)
	}

	return handler.eventQueue
}
