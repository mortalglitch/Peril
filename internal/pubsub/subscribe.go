package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
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

func workerRoutine[T any](messages <-chan amqp.Delivery, handler func(T) AckType) {
	for msg := range messages {
		var singleMsg T
		messageResult := json.Unmarshal(msg.Body, &singleMsg)
		if messageResult != nil {
			log.Printf("Error consuming message from MQ: %v", messageResult)
			return
		}

		ackType := handler(singleMsg)

		if ackType == Ack {
			msg.Ack(false)
		} else if ackType == NackRequeue {
			msg.Nack(false, true)
		} else if ackType == NackDiscard {
			msg.Nack(false, false)
		}
		//log.Printf("AckType Sent: %s", ackType)
	}
}
