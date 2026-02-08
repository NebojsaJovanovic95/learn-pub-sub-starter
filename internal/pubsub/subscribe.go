package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T),
) error {
	
	ch, _, err := ch.consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func () {
		for d := range msgs {
			var val T
			if err := json.Unmarshal(d.Body, &val); err != nil {
				log.Println("Failed to unmarshal message:", err)
				d.Nack(false, false)
				continue
			}

			handler(val)

			d.Ack(false)
		}
	} ()

	return nil
}
