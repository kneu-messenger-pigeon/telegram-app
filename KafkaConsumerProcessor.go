package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/segmentio/kafka-go"
	"io"
	"sync"
	"time"
)

type KafkaConsumerProcessor struct {
	out             io.Writer
	reader          events.ReaderInterface
	handler         EventHandlerInterface
	commitThreshold int
	disabled        bool
}

const defaultCommitThreshold = 10

func (processor *KafkaConsumerProcessor) Execute(ctx context.Context, wg *sync.WaitGroup) {
	var err error
	var message kafka.Message
	var messagesToCommit []kafka.Message
	var fetchContext = ctx
	var fetchContextCancel func()

	expectedMessageKey := processor.handler.GetExpectedMessageKey()
	event := processor.handler.GetExpectedEventType()

	if processor.commitThreshold == 0 {
		processor.commitThreshold = defaultCommitThreshold
	}

	if expectedMessageKey == "" || event == nil || processor.disabled {
		wg.Done()
		return
	}

	_, _ = fmt.Fprintf(processor.out, "Consuming for %T started \n", processor.handler)

	for ctx.Err() == nil {
		message, err = processor.reader.FetchMessage(fetchContext)
		if err == nil && expectedMessageKey == string(message.Key) {
			err = json.Unmarshal(message.Value, &event)
			if err == nil {
				err = processor.handler.Handle(event)
			}
		}
		if err == nil {
			if len(messagesToCommit) == 0 {
				// set context with timeout to make sure that every 60 seconds we Execute Commit
				fetchContext, fetchContextCancel = context.WithTimeout(ctx, time.Second*60)
			}
			messagesToCommit = append(messagesToCommit, message)
		}

		if len(messagesToCommit) != 0 && (len(messagesToCommit) >= processor.commitThreshold || fetchContext.Err() != nil) {
			// revert context with time to usual
			fetchContext = ctx
			err = processor.handler.Commit()
			if err == nil {
				err = processor.reader.CommitMessages(context.Background(), messagesToCommit...)
			}
			_, _ = fmt.Fprintf(processor.out, "%T Commit %d messages (err: %v) \n", processor.handler, len(messagesToCommit), err)
			if err == nil {
				messagesToCommit = []kafka.Message{}
			}
		}

		if err != nil && err != context.Canceled {
			_, _ = fmt.Fprintf(processor.out, "%T error: %v \n", processor.handler, err)
		}
	}
	if fetchContextCancel != nil {
		fetchContextCancel()
	}

	_, _ = fmt.Fprintf(processor.out, "ScoreChangedEventHandler done \n")

	wg.Done()
}

func (processor *KafkaConsumerProcessor) Disable() {
	processor.disabled = true
}
