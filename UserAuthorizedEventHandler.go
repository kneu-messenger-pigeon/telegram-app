package main

import (
	"fmt"
	"github.com/kneu-messenger-pigeon/events"
	"io"
)

type UserAuthorizedEventHandler struct {
	clientName       string
	repository       UserRepositoryInterface
	out              io.Writer
	serviceContainer *ServiceContainer
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

		if err == nil && handler.serviceContainer != nil && handler.serviceContainer.ClientController != nil {
			go handler.callControllerAction(event)
		}
	}

	return err
}

func (handler *UserAuthorizedEventHandler) callControllerAction(event *events.UserAuthorizedEvent) {
	err := handler.serviceContainer.ClientController.UserAuthorizedAction(event)
	if err != nil {
		_, _ = fmt.Fprintf(handler.out, "UserAuthorizedAction return error: %v", err)
	}
}
