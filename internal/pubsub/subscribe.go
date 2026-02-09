package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType, "dead_letter_exchange")
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer name (auto-generated)
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var payload T
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Println("Error unmarshaling message:", err)
				d.Ack(false)
				continue
			}
			ack := handler(payload)
			switch ack {
			case Ack:
				d.Ack(false)
				log.Println("Ack called")
			case NackRequeue:
				d.Nack(false, true)
				log.Println("Nack called with requeue")
			case NackDiscard:
				d.Nack(false, false)
				log.Println("Nack called with discard")
			default:
				continue
			}
		}
	}()

	return nil
}
