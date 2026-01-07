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

	newState := gamelogic.NewGameState(usernameString)

	subscribeSuccess := pubsub.SubscribeJSON(newConnection, routing.ExchangePerilDirect, routing.PauseKey+"."+usernameString, routing.PauseKey, pubsub.Transient, handlerPause(newState))
	if subscribeSuccess != nil {
		log.Fatalf("Error with subscribe process: %v", subscribeSuccess)
	}

	for {
		result := gamelogic.GetInput()
		if len(result) == 0 {
			continue
		} else if result[0] == "spawn" {
			newState.CommandSpawn(result)
		} else if result[0] == "move" {
			newState.CommandMove(result)
		} else if result[0] == "status" {
			newState.CommandStatus()
		} else if result[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if result[0] == "spam" {
			log.Println("Spamming not allowed yet!")
		} else if result[0] == "quit" {
			gamelogic.PrintQuit()
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
