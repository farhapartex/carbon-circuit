package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	EventIDHeader    = "event_id"
	DeadLetterSuffix = ".dlq"
)

type Delivery struct {
	Topic   string
	Key     []byte
	Value   []byte
	EventID string
}

type Handler func(ctx context.Context, delivery Delivery) error

type DeadLetter interface {
	Publish(ctx context.Context, messages ...Message) error
}

type ConsumerOptions struct {
	Brokers    []string
	Group      string
	Topics     []string
	Logger     *slog.Logger
	Handle     Handler
	DeadLetter DeadLetter
}

type Consumer struct {
	client     *kgo.Client
	logger     *slog.Logger
	handle     Handler
	deadLetter DeadLetter
}

func NewConsumer(options ConsumerOptions) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(options.Brokers...),
		kgo.ConsumerGroup(options.Group),
		kgo.ConsumeTopics(options.Topics...),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	return &Consumer{
		client:     client,
		logger:     options.Logger,
		handle:     options.Handle,
		deadLetter: options.DeadLetter,
	}, nil
}

func (c *Consumer) Close() { c.client.Close() }

func (c *Consumer) Run(ctx context.Context) {
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			if !errors.Is(err, context.Canceled) {
				c.logger.Error("fetch failed",
					slog.String("topic", topic),
					slog.Int("partition", int(partition)),
					slog.Any("error", err),
				)
			}
		})

		blocked := false

		fetches.EachRecord(func(record *kgo.Record) {
			delivery := deliveryFrom(record)

			err := c.handle(ctx, delivery)
			if err == nil {
				return
			}

			c.logger.Error("handling event failed",
				slog.String("topic", delivery.Topic),
				slog.String("event_id", delivery.EventID),
				slog.Any("error", err),
			)

			if !c.divert(ctx, delivery, err) {
				blocked = true
			}
		})

		if blocked {
			continue
		}

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.logger.Error("committing offsets failed", slog.Any("error", err))
		}
	}
}

func deliveryFrom(record *kgo.Record) Delivery {
	delivery := Delivery{
		Topic: record.Topic,
		Key:   record.Key,
		Value: record.Value,
	}

	for _, header := range record.Headers {
		if header.Key == EventIDHeader {
			delivery.EventID = string(header.Value)
		}
	}

	return delivery
}

func (c *Consumer) divert(
	ctx context.Context,
	delivery Delivery,
	cause error,
) bool {
	if c.deadLetter == nil {
		return false
	}

	err := c.deadLetter.Publish(ctx, Message{
		Topic: delivery.Topic + DeadLetterSuffix,
		Key:   delivery.Key,
		Value: delivery.Value,
		Headers: map[string]string{
			EventIDHeader:    delivery.EventID,
			"origin_topic":   delivery.Topic,
			"failure_reason": cause.Error(),
		},
	})
	if err != nil {
		c.logger.Error("dead lettering failed, holding the partition",
			slog.String("topic", delivery.Topic),
			slog.String("event_id", delivery.EventID),
			slog.Any("error", err),
		)
		return false
	}

	c.logger.Warn("event dead lettered",
		slog.String("topic", delivery.Topic),
		slog.String("event_id", delivery.EventID),
	)

	return true
}
