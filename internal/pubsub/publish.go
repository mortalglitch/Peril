package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return err
	}

	publishError := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	if publishError != nil {
		log.Printf("Error publishing data: %s", publishError)
		return publishError
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var encByte bytes.Buffer
	enc := gob.NewEncoder(&encByte)
	err := enc.Encode(val)
	if err != nil {
		log.Printf("Error encoding gob data: %s", err)
		return err
	}

	publishError := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        encByte.Bytes(),
	})
	if publishError != nil {
		log.Printf("Error publishing data: %s", publishError)
		return publishError
	}

	return nil
}
