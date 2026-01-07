package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {
	channel, boundQueue, binderr := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if binderr != nil {
		log.Fatalf("Error binding to channel and queue: %s", binderr)
	}

	msg, err := channel.Consume(boundQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("Error consuming message from MQ: %v", err)
		return err
	}

	go workerRoutine(msg, handler)

	return nil
}

func workerRoutine[T any](messages <-chan amqp.Delivery, handler func(T)) {
	for msg := range messages {
		var singleMsg T
		messageResult := json.Unmarshal(msg.Body, &singleMsg)
		if messageResult != nil {
			log.Printf("Error consuming message from MQ: %v", messageResult)
			return
		}
		handler(singleMsg)
		msg.Ack(false)
	}
}
