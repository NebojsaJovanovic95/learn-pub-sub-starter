package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	deadLetterExchange string,
) (*amqp.Channel, amqp.Queue, error) {

	args := amqp.Table{}
	if deadLetterExchange != "" {
		args["x-dead-letter-exchange"] = deadLetterExchange
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := queueType == Durable
	autoDelete := queueType == Transient
	exclusive := queueType == Transient

	q, err := ch.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		args,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err := ch.QueueBind(
		q.Name,
		key,
		exchange,
		false,
		nil,
	);
	err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
