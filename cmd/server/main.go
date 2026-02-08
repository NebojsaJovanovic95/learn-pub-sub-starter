package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/NebojsaJovanovic95/learn-pub-sub-starter/internal/pubsub"
	"github.com/NebojsaJovanovic95/learn-pub-sub-starter/internal/routing"
	"github.com/NebojsaJovanovic95/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		routing.ExchangePerilTopic,
		"topic",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare peril_topic exchange:", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		"game_logs",
		routing.GameLogSlug + ".*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatal("Failed to declare/bind game_logs queue:", err)
	}

	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

	// REPL loop
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Println("Error publishing pause:", err)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Println("Error publishing resume:", err)
			}
		case "quit":
			fmt.Println("Exiting server...")
			return
		default:
			fmt.Println("Unknown command:", words[0])
		}
	}
}
