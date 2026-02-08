package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	username, err := gamelogic.ClientWelcome()
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

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nClient exiting...")
		os.Exit(0)
	}()

	// REPL loop
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			id := state.CommandSpawn(words)
			fmt.Printf("Spawned unit with ID %d\n", id)
		case "move":
			mv, err := state.CommandMove(words)
			if err != nil {
				fmt.Println("Error moving unit:", err)
			} else {
				fmt.Printf("Move successful! Moved %d unit(s) to %s\n", len(mv.Units), mv.ToLocation)
			}
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
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
			fmt.Println("Exiting client...")
			return
		default:
			fmt.Println("Unknown command:", words[0])
		}
	}
}
