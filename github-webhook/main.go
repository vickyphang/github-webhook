package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
)

var webhookSecret string = os.Getenv("WEBHOOK_SECRET")
var amqpURL string = os.Getenv("AMQP_URL")
var queueName string = os.Getenv("QUEUE_NAME")
var port string = os.Getenv("LISTEN_PORT")

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

func verifySignature(payload []byte, signature string) bool {
	mac := hmac.New(sha1.New, []byte(webhookSecret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha1=" + hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Hub-Signature")
	if !verifySignature(body, signature) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	var event PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "could not parse JSON", http.StatusInternalServerError)
		return
	}

	// Send the event to RabbitMQ
	err = sendToQueue(body)
	if err != nil {
		http.Error(w, "could not send to queue", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func sendToQueue(body []byte) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
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
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
