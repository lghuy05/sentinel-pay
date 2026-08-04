package kafkautil

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

const DLTHeaderError = "x-error"

func NewDLTWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
}

func PublishDLT(ctx context.Context, writer *kafka.Writer, message kafka.Message, cause error) error {
	if writer == nil {
		return fmt.Errorf("dlt writer is nil")
	}
	return writer.WriteMessages(ctx, DLTMessage(message, cause))
}

func DLTMessage(message kafka.Message, cause error) kafka.Message {
	headers := append([]kafka.Header{}, message.Headers...)
	if cause != nil {
		headers = append(headers, kafka.Header{Key: DLTHeaderError, Value: []byte(cause.Error())})
	}
	return kafka.Message{
		Topic:   message.Topic + ".DLT",
		Key:     message.Key,
		Value:   message.Value,
		Headers: headers,
	}
}
