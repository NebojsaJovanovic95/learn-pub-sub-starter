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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) (pubsub.AckType) {
	return func(ps routing.PlayingState) (pubsub.AckType) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(mv gamelogic.ArmyMove) (pubsub.AckType) {
	return func(mv gamelogic.ArmyMove) (pubsub.AckType) {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		switch outcome {
		case gamelogic.MoveOutComeSafe, gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	state := gamelogic.NewGameState(username)

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(state),
	)
	if err != nil {
		log.Fatal("Failed to subscribe to pause messages:", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix, username),
		pubsub.Transient,
		handlerMove(state),
	)
	if err != nil {
		log.Fatal("Failed to subscribe to move messages:", err)
	}

	fmt.Println("Client running. Press Ctrl+C to exit.")

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
				continue
			}

			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				mv,
			)
			if err != nil {
				log.Println("Failed to publish move:", err)
				continue
			}

			fmt.Println("Move published successfully")
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
