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

	pauseSubSuccess := pubsub.SubscribeJSON(newConnection, routing.ExchangePerilDirect, routing.PauseKey+"."+usernameString, routing.PauseKey, pubsub.Transient, handlerPause(newState))
	if pauseSubSuccess != nil {
		log.Fatalf("Error with subscribe process: %v", pauseSubSuccess)
	}

	moveSubSuccess := pubsub.SubscribeJSON(newConnection, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+usernameString, routing.ArmyMovesPrefix+".*", pubsub.Transient, handlerMove(newState))
	if moveSubSuccess != nil {
		log.Fatalf("Error getting moves from MQ %v", moveSubSuccess)
	}

	rabbitChannel, err := newConnection.Channel()
	if err != nil {
		log.Fatalf("Error connecting to channel: %s", binderr)
	}

	for {
		result := gamelogic.GetInput()
		if len(result) == 0 {
			continue
		} else if result[0] == "spawn" {
			newState.CommandSpawn(result)
		} else if result[0] == "move" {
			armyMove, err := newState.CommandMove(result)
			if err != nil {
				log.Println("Trouble with move: ", err)
			}
			pubsub.PublishJSON(rabbitChannel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+usernameString, armyMove)
			log.Println("Success published move.")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(am)
		if moveOutcome == gamelogic.MoveOutComeSafe || moveOutcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}
