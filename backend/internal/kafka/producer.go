package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string
}

type Producer struct {
	client *kgo.Client
}

type ProducerOptions struct {
	Brokers          []string
	AllowTopicCreate bool
}

func NewProducer(options ProducerOptions) (*Producer, error) {
	settings := []kgo.Opt{
		kgo.SeedBrokers(options.Brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerLinger(0),
	}

	if options.AllowTopicCreate {
		settings = append(settings, kgo.AllowAutoTopicCreation())
	}

	client, err := kgo.NewClient(settings...)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{client: client}, nil
}

func (p *Producer) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}

func (p *Producer) Close() { p.client.Close() }

func (p *Producer) Publish(ctx context.Context, messages ...Message) error {
	records := make([]*kgo.Record, 0, len(messages))
	for _, message := range messages {
		records = append(records, recordFrom(message))
	}

	for _, result := range p.client.ProduceSync(ctx, records...) {
		if result.Err != nil {
			return fmt.Errorf("publish to %s: %w", result.Record.Topic, result.Err)
		}
	}

	return nil
}

func recordFrom(message Message) *kgo.Record {
	headers := make([]kgo.RecordHeader, 0, len(message.Headers))
	for key, value := range message.Headers {
		headers = append(headers, kgo.RecordHeader{Key: key, Value: []byte(value)})
	}

	return &kgo.Record{
		Topic:   message.Topic,
		Key:     message.Key,
		Value:   message.Value,
		Headers: headers,
	}
}
