package main

import "github.com/kneu-messenger-pigeon/events"

type ClientControllerInterface interface {
	ExecutorInterface
	ScoreChangedAction(event *events.ScoreChangedEvent) error
	UserAuthorizedAction(event *events.UserAuthorizedEvent) error
}
