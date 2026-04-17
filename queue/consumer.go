package queue

import (
    "encoding/json"
    "log"
    "os"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/YoussufAlgonquin/bestbuy-makeline-service/db"
    "github.com/YoussufAlgonquin/bestbuy-makeline-service/models"
)

const QueueName = "orders"

func StartConsuming() {
    uri := os.Getenv("RABBITMQ_URI")
    if uri == "" {
        log.Fatal("RABBITMQ_URI not set")
    }

    var conn *amqp.Connection
    var err error
    for i := 0; i < 5; i++ {
        conn, err = amqp.Dial(uri)
        if err == nil {
            break
        }
        log.Printf("RabbitMQ connection attempt %d failed: %v", i+1, err)
        time.Sleep(3 * time.Second)
    }
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open channel: %v", err)
    }

    // Declare the queue — idempotent, safe to call even if queue already exists
    _, err = ch.QueueDeclare(QueueName, true, false, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }

    // Prefetch 1: only give this consumer one message at a time.
    // Without this, RabbitMQ could flood the consumer with all pending messages.
    ch.Qos(1, 0, false)

    msgs, err := ch.Consume(QueueName, "", false, false, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to register consumer: %v", err)
    }

    log.Println("makeline-service listening for orders...")

    // Process messages in a goroutine so the health server can still run
    go func() {
        for msg := range msgs {
            processOrder(msg)
        }
    }()
}

func processOrder(msg amqp.Delivery) {
    var order models.QueueMessage
    if err := json.Unmarshal(msg.Body, &order); err != nil {
        log.Printf("Failed to parse message: %v", err)
        msg.Nack(false, false) // Reject message, don't requeue (malformed)
        return
    }

    log.Printf("Processing order %s for customer %s", order.OrderID, order.CustomerID)

    // Mark as processing
    if err := db.UpdateOrderStatus(order.OrderID, "processing"); err != nil {
        log.Printf("Failed to update order status to processing: %v", err)
        msg.Nack(false, true) // Requeue — transient DB error, try again
        return
    }

    // Simulate warehouse pick-and-pack time
    time.Sleep(3 * time.Second)

    // Mark as complete
    if err := db.UpdateOrderStatus(order.OrderID, "complete"); err != nil {
        log.Printf("Failed to update order status to complete: %v", err)
        msg.Nack(false, true)
        return
    }

    log.Printf("Order %s complete", order.OrderID)
    msg.Ack(false) // Acknowledge — tells RabbitMQ to remove the message from the queue
}