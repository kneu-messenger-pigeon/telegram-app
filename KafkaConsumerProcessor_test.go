package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

func TestKafkaToRedisConnector(t *testing.T) {
	matchContext := mock.MatchedBy(func(ctx context.Context) bool { return true })

	t.Run("Two iteration and Commit", func(t *testing.T) {
		out := &bytes.Buffer{}

		event := events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: "999",
			StudentId:    999,
		}

		payload, _ := json.Marshal(event)
		message := kafka.Message{
			Key:   []byte(events.UserAuthorizedEventName),
			Value: payload,
		}

		ctx, cancel := context.WithCancel(context.Background())

		handler := NewMockEventHandlerInterface(t)
		handler.On("GetExpectedMessageKey").Return(events.UserAuthorizedEventName)
		handler.On("GetExpectedEventType").Return(&events.UserAuthorizedEvent{})
		handler.On("Handle", &event).Once().Return(nil)
		handler.On("Commit").Return(nil)

		reader := events.NewMockReaderInterface(t)

		reader.On("FetchMessage", matchContext).Once().Return(message, nil)
		reader.On("FetchMessage", matchContext).Once().Return(func(_ context.Context) kafka.Message {
			cancel()
			return kafka.Message{}
		}, nil)

		reader.On("CommitMessages", matchContext, message, kafka.Message{}).Return(nil)

		processor := KafkaConsumerProcessor{
			out:     out,
			reader:  reader,
			handler: handler,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		processor.Execute(ctx, &wg)
	})

	t.Run("Emulate write error", func(t *testing.T) {
		expectedError := errors.New("expected error")

		out := &bytes.Buffer{}

		event := events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: "999",
			StudentId:    999,
		}

		payload, _ := json.Marshal(event)
		message := kafka.Message{
			Key:   []byte(events.UserAuthorizedEventName),
			Value: payload,
		}

		ctx, cancel := context.WithCancel(context.Background())

		handler := NewMockEventHandlerInterface(t)
		handler.On("Handle", &event).Once().Return(nil)
		handler.On("GetExpectedMessageKey").Return(events.UserAuthorizedEventName)
		handler.On("GetExpectedEventType").Return(&events.UserAuthorizedEvent{})
		handler.On("Commit").Return(expectedError)

		reader := events.NewMockReaderInterface(t)

		reader.On("FetchMessage", matchContext).Once().Return(func(_ context.Context) kafka.Message {
			cancel()
			return message
		}, nil)

		connector := KafkaConsumerProcessor{
			out:             out,
			commitThreshold: 1,
			reader:          reader,
			handler:         handler,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		connector.Execute(ctx, &wg)

		assert.Contains(t, out.String(), expectedError.Error())
	})

	t.Run("Emulate writer init error", func(t *testing.T) {
		out := &bytes.Buffer{}

		handler := NewMockEventHandlerInterface(t)
		handler.On("GetExpectedMessageKey").Return("")
		handler.On("GetExpectedEventType").Return(&events.UserAuthorizedEvent{})

		reader := events.NewMockReaderInterface(t)

		connector := KafkaConsumerProcessor{
			out:     out,
			reader:  reader,
			handler: handler,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		ctx, fetchCancel := context.WithTimeout(context.Background(), time.Millisecond)
		connector.Execute(ctx, &wg)
		fetchCancel()

		reader.AssertNotCalled(t, "FetchMessage")
		reader.AssertNotCalled(t, "CommitMessages")
		handler.AssertNotCalled(t, "Handle")

		assert.Empty(t, out.String())
	})
}
