package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	newChannel, err := conn.Channel()
	if err != nil {
		log.Println("Error when building new channel. ", err)
		return nil, amqp.Queue{}, err
	}

	var durable bool
	var autodelete bool
	var exclusive bool

	if queueType == Durable {
		durable = true
		autodelete = false
		exclusive = false
	}
	if queueType == Transient {
		durable = false
		autodelete = true
		exclusive = true
	}

	// Dead-Letter exchange config
	deadLetterExchange := "peril_dlx"

	params := amqp.Table{
		"x-dead-letter-exchange": deadLetterExchange,
	}

	newQueue, err := newChannel.QueueDeclare(queueName, durable, autodelete, exclusive, false, params)
	if err != nil {
		log.Println("Error declaring queue. ", err)
		return nil, amqp.Queue{}, err
	}

	newBoundResult := newChannel.QueueBind(queueName, key, exchange, false, nil)
	if newBoundResult != nil {
		if err != nil {
			log.Println("Error binding queue. ", err)
			return nil, amqp.Queue{}, err
		}
	}

	return newChannel, newQueue, nil
}
