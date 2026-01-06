package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	rmqConnection := "amqp://guest:guest@localhost:5672/"

	newConnection, err := amqp.Dial(rmqConnection)
	if err != nil {
		log.Fatalf("Trouble dialing rabbitMQ: %v", err)
	}
	defer newConnection.Close()

	usernameString, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Failed to build username: %s", err)
	}

	_, _, binderr := pubsub.DeclareAndBind(newConnection, routing.ExchangePerilDirect, routing.PauseKey+"."+usernameString, routing.PauseKey, pubsub.Transient)
	if binderr != nil {
		log.Fatalf("Error binding to channel and queue: %s", binderr)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interrupt := <-signalChan
	fmt.Println("Server shutting down: ", interrupt)
	newConnection.Close()

}
