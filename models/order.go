package models

// Order mirrors the structure saved by order-service in MongoDB.
// We only need the fields we care about — Go structs are flexible like this.
type Order struct {
    ID         string     `bson:"_id" json:"orderId"`
    CustomerID string     `bson:"customerId" json:"customerId"`
    Status     string     `bson:"status" json:"status"`
}

// QueueMessage is what order-service publishes to RabbitMQ
type QueueMessage struct {
    OrderID    string      `json:"orderId"`
    CustomerID string      `json:"customerId"`
    TotalPrice float64     `json:"totalPrice"`
}