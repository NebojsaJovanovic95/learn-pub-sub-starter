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

	username := gamelogic.ClientWelcome()
	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	fmt.Println("Client running. Press Ctrl+C to exit.")

	state := gamelogic.NewGameState(username)

	for ;; {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			id, err := state.CommandSpawn(words)
			if err != nil {
				fmt.Println("Error spawning unit:", err)
			} else {
				fmt.Println("Spawned unit with ID %d\n", err)
			}
		case "move":
			err := state.CommandMove(words)
			if err != nil {
				fmt.Println("Error moving unit:", err)
			} else {
				fmt.Println("Move successfully!")
			}
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "pause":
			fmt.Println("sending pause message...")
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
				log.Println("Error publishing resume...", err)
			}
		case "quit":
			fmt.Println("Exiting server...")
			return
		default:
			fmt.Println("Unknown commands:", words[0])
		}
	}
}
