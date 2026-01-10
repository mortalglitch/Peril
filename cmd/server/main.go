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
	fmt.Println("Starting Peril server...")
	rmqConnection := "amqp://guest:guest@localhost:5672/"

	newConnection, err := amqp.Dial(rmqConnection)
	if err != nil {
		log.Fatalf("Trouble dialing rabbitMQ: %v", err)
	}
	defer newConnection.Close()

	fmt.Println("Server connection successful")
	gamelogic.PrintServerHelp()

	channel, err := newConnection.Channel()
	if err != nil {
		log.Fatalf("Trouble creating connection channel: %v", err)
	}

	_, _, binderr := pubsub.DeclareAndBind(newConnection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable)
	if binderr != nil {
		log.Fatalf("Error binding to channel and queue: %s", binderr)
	}

	gameLogSubSuccess := pubsub.SubscribeGob(newConnection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable, handlerGameLogs(channel))
	if gameLogSubSuccess != nil {
		log.Fatalf("Error getting moves from MQ %v", gameLogSubSuccess)
	}

	for {
		result := gamelogic.GetInput()
		if len(result) == 0 {
			continue
		} else if result[0] == "pause" {
			log.Println("Sending pause message.")
			messageSent := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if messageSent != nil {
				log.Fatalf("Error sending message to RabbitMQ: %v", err)
			}
		} else if result[0] == "resume" {
			log.Println("Sending resume message.")
			messageSent := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if messageSent != nil {
				log.Fatalf("Error sending message to RabbitMQ: %v", err)
			}
		} else if result[0] == "quit" {
			log.Println("Exiting game.")
			break
		} else {
			log.Println("Sorry I do not understand the request.")
		}

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interrupt := <-signalChan
	fmt.Println("Server shutting down: ", interrupt)
	newConnection.Close()
}

func handlerGameLogs(rabbitChannel *amqp.Channel) func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		gameLogSuccess := gamelogic.WriteLog(gl)
		if gameLogSuccess != nil {
			log.Printf("Error reading logs from RabbitMQ: %v", gameLogSuccess)
			return pubsub.NackRequeue
		}

		return pubsub.NackDiscard
	}
}
