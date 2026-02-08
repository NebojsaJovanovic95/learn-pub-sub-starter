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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
