package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var amqpURL string = os.Getenv("AMQP_URL")
var queueName string = os.Getenv("QUEUE_NAME")

type Commit struct {
	Message string `json:"message"`
	Author  struct {
		Name string `json:"name"`
	} `json:"author"`
}

type PushEvent struct {
	Ref     string   `json:"ref"`
	Commits []Commit `json:"commits"`
}

func main() {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var event PushEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("could not parse JSON: %v", err)
				continue
			}

			// Process the push event data here
			fmt.Printf("Received push event: %+v\n", event)
			for _, commit := range event.Commits {
				fmt.Printf("Commit by %s: %s\n", commit.Author.Name, commit.Message)
			}
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
