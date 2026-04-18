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

    // Run in a goroutine so the health server starts immediately.
    // Loop forever so a dropped RabbitMQ connection triggers a full reconnect.
    go func() {
        for {
            if err := consume(uri); err != nil {
                log.Printf("Consumer exited with error: %v — reconnecting in 5s", err)
            } else {
                log.Println("Consumer channel closed — reconnecting in 5s")
            }
            time.Sleep(5 * time.Second)
        }
    }()
}

func consume(uri string) error {
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
        return err
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        return err
    }
    defer ch.Close()

    // Declare the queue — idempotent, safe to call even if queue already exists
    _, err = ch.QueueDeclare(QueueName, true, false, false, false, nil)
    if err != nil {
        return err
    }

    // Prefetch 1: only give this consumer one message at a time.
    ch.Qos(1, 0, false)

    msgs, err := ch.Consume(QueueName, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    log.Println("makeline-service listening for orders...")

    for msg := range msgs {
        processOrder(msg)
    }

    return nil
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

    for _, item := range order.LineItems {
        if err := db.DecrementProductStock(item.ProductID, item.Quantity); err != nil {
            log.Printf("Failed to decrement stock for product %s: %v", item.ProductID, err)
        }
    }

    log.Printf("Order %s complete", order.OrderID)
    msg.Ack(false) // Acknowledge — tells RabbitMQ to remove the message from the queue
}