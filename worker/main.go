package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func processTask(message amqp.Delivery) {
	// Insert your task processing logic here
	fmt.Printf("Received a message: %s\n", message.Body)

	// Acknowledge the message after processing
	message.Ack(false)
}

func main() {
	// Connects to RabbitMQ server
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Creates a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declares a queue for messages
	q, err := ch.QueueDeclare(
		"task_queue", // name of the queue
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Receive messages from the queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-acknowledge
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	// Listen for messages and process them
	go func() {
		for d := range msgs {
			processTask(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
