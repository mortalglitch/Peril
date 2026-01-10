package pubsub

import (
	"bytes"
	"encoding/gob"
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

	go workerJSONRoutine(msg, handler)

	return nil
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	channel, boundQueue, binderr := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if binderr != nil {
		log.Fatalf("Error binding to channel and queue: %s", binderr)
	}

	err := channel.Qos(10, 0, false)
	if err != nil {
		log.Printf("Error setting channel parameters: %v", err)
		return err
	}

	msg, err := channel.Consume(boundQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("Error consuming message from MQ: %v", err)
		return err
	}

	go workerGobRoutine(msg, handler)

	return nil
}

func workerJSONRoutine[T any](messages <-chan amqp.Delivery, handler func(T) AckType) {
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

func workerGobRoutine[T any](messages <-chan amqp.Delivery, handler func(T) AckType) {
	for msg := range messages {
		var singleMessage T
		err := gobUnmarshaller(msg.Body, &singleMessage)
		if err != nil {
			log.Printf("Error consuming message from MQ: %v", err)
			return
		}

		ackType := handler(singleMessage)

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

func gobUnmarshaller[T any](toDecode []byte, v *T) error {
	bufferData := bytes.NewBuffer(toDecode)
	dec := gob.NewDecoder(bufferData)
	err := dec.Decode(v)
	return err
}
