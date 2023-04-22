package main

import (
	"github.com/kneu-messenger-pigeon/events"
)

type UserAuthorizedEventHandler struct {
	repository UserRepositoryInterface
	clientName string
	eventQueue chan *events.UserAuthorizedEvent
}

func (handler *UserAuthorizedEventHandler) GetExpectedMessageKey() string {
	return events.UserAuthorizedEventName
}

func (handler *UserAuthorizedEventHandler) GetExpectedEventType() any {
	return &events.UserAuthorizedEvent{}
}

func (handler *UserAuthorizedEventHandler) Commit() error {
	return handler.repository.Commit()
}

func (handler *UserAuthorizedEventHandler) Handle(s any) (err error) {
	event := s.(*events.UserAuthorizedEvent)
	if event.Client == handler.clientName {
		err = handler.repository.SaveUser(event.ClientUserId, &Student{
			Id:         uint32(event.StudentId),
			LastName:   event.LastName,
			FirstName:  event.FirstName,
			MiddleName: event.MiddleName,
			Gender:     Student_GenderType(event.Gender),
		})
		if handler.eventQueue != nil {
			handler.eventQueue <- event
		}
	}

	return err
}

func (handler *UserAuthorizedEventHandler) GetEventQueue() <-chan *events.UserAuthorizedEvent {
	if handler.eventQueue == nil {
		handler.eventQueue = make(chan *events.UserAuthorizedEvent)
	}

	return handler.eventQueue
}
