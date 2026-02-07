package main

import (
	"github.com/rabbitmq/amqp091-go"
	"fmt"
)

func main() {
	connString := "amqp://guest:guest@localhost:5672/"
	conn := amqp.Dial(connString)
	defer conn.Close()
	fmt.Println("Starting Peril server...")
}
