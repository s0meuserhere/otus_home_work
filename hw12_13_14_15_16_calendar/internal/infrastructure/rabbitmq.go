package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/config"
)

const (
	rabbitExchangeKind = "direct"
	// Суффиксы для exchange и очереди недоставленных сообщений.
	rabbitDeadLetterExchangeSuffix = ".dlx"
	rabbitDeadLetterQueueSuffix    = ".dlq"
)

// rabbitTopology - структуры, которые создаются при подключении.
type rabbitTopology struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

// RabbitClient подключается к RabbitMQ, отправляет и читает сообщения.
type RabbitClient struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	topology rabbitTopology
	closed   chan error
}

// NewRabbitClient подключается к RabbitMQ и создаёт exchange и очередь, если их ещё нет.
func NewRabbitClient(cfg config.RabbitConf) (*RabbitClient, error) {
	url, err := cfg.URL()
	if err != nil {
		return nil, fmt.Errorf("rabbit url: %w", err)
	}

	topology := rabbitTopology{
		Exchange:   cfg.Exchange,
		Queue:      cfg.Queue,
		RoutingKey: cfg.RoutingKey,
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("open channel: %w", err)
	}

	c := &RabbitClient{
		conn:     conn,
		ch:       ch,
		topology: topology,
		closed:   make(chan error, 1),
	}

	if err := c.setup(); err != nil {
		_ = c.Close()

		return nil, err
	}

	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		if amqpErr, ok := <-connClosed; ok && amqpErr != nil {
			c.closed <- fmt.Errorf("connection closed: %w", amqpErr)
		}
		close(c.closed)
	}()

	return c, nil
}

// setup создаёт основную очередь и очередь недоставленных сообщений.
// Отклонённое сообщение брокер перекладывает в очередь недоставленных, а не удаляет.
func (c *RabbitClient) setup() error {
	t := c.topology
	dlx := t.Exchange + rabbitDeadLetterExchangeSuffix

	if err := c.declare(dlx, t.Queue+rabbitDeadLetterQueueSuffix, t.RoutingKey, nil); err != nil {
		return err
	}

	if err := c.declare(t.Exchange, t.Queue, t.RoutingKey, amqp.Table{"x-dead-letter-exchange": dlx}); err != nil {
		return err
	}

	// Publish ждёт подтверждения брокера.
	if err := c.ch.Confirm(false); err != nil {
		return fmt.Errorf("enable publisher confirms: %w", err)
	}

	return nil
}

// declare создаёт exchange и очередь, если их ещё нет, и связывает их по routingKey.
func (c *RabbitClient) declare(exchange, queue, routingKey string, queueArgs amqp.Table) error {
	if err := c.ch.ExchangeDeclare(exchange, rabbitExchangeKind, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", exchange, err)
	}

	if _, err := c.ch.QueueDeclare(queue, true, false, false, false, queueArgs); err != nil {
		return fmt.Errorf("declare queue %q: %w", queue, err)
	}

	if err := c.ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue %q to %q: %w", queue, exchange, err)
	}

	return nil
}

// Wait ждёт отмены ctx и возвращает ошибку, если соединение разорвано.
func (c *RabbitClient) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err, ok := <-c.closed:
		if ok {
			return err
		}

		return nil
	}
}

// Publish публикует сообщение и ждёт подтверждения брокера.
func (c *RabbitClient) Publish(ctx context.Context, body []byte) error {
	confirm, err := c.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		c.topology.Exchange,
		c.topology.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("wait publish confirm: %w", err)
	}

	if !acked {
		return errors.New("publish: message nacked by broker")
	}

	return nil
}

// Consume передаёт сообщения в handle, а сообщения с ошибкой перекладывает в очередь недоставленных.
func (c *RabbitClient) Consume(
	ctx context.Context,
	prefetch int,
	handle func(ctx context.Context, body []byte) error,
) error {
	if err := c.ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	queue := c.topology.Queue
	deliveries, err := c.ch.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %q: %w", queue, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				select {
				case <-ctx.Done():
					return nil
				default:
					return errors.New("deliveries channel closed")
				}
			}

			if err := handle(ctx, d.Body); err != nil {
				if nackErr := d.Nack(false, false); nackErr != nil {
					return fmt.Errorf("nack: %w", nackErr)
				}

				continue
			}

			if err := d.Ack(false); err != nil {
				return fmt.Errorf("ack: %w", err)
			}
		}
	}
}

// Close закрывает соединение вместе с каналом.
func (c *RabbitClient) Close() error {
	if err := c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}
