package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	rmqConnection := "amqp://guest:guest@localhost:5672/"

	newConnection, err := amqp.Dial(rmqConnection)
	if err != nil {
		log.Fatalf("Trouble dialing rabbitMQ: %v", err)
	}
	defer newConnection.Close()

	fmt.Println("Server connection successful")

	channel, err := newConnection.Channel()
	if err != nil {
		log.Fatalf("Trouble creating connection channel: %v", err)
	}

	messageSent := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	if messageSent != nil {
		log.Fatalf("Error sending message to RabbitMQ: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interrupt := <-signalChan
	fmt.Println("Server shutting down: ", interrupt)
	newConnection.Close()
}
