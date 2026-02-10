package pubsub

import (
	"bytes"
	"encoding/gob"
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
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
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

func GobUnmarshal[T any](data []byte, payload *T) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(payload)
}

func Subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte, *T) error,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	if err := ch.Qos(10, 0, true); err != nil {
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
		for msg := range msgs {
			var payload T
			if err := unmarshaller(msg.Body, &payload); err != nil {
				log.Println("Error unmarshaling message:", err)
				msg.Ack(false)
				continue
			}
			ack := handler(payload)
			switch ack {
			case Ack:
				msg.Ack(false)
				log.Println("Ack called")
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("Nack called with requeue")
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("Nack called with discard")
			default:
				continue
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	err := Subscribe(conn, exchange, queueName, key, queueType, handler, GobUnmarshal)
	if err != nil {
		return err
	}
	return nil
}
