# BestBuy Makeline Service

A Go service that reads orders from a RabbitMQ queue and stores them in MongoDB. It runs as a background processor and exposes a health endpoint.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |

## How It Works

1. Connects to RabbitMQ and subscribes to the orders queue
2. For each incoming order message, writes the order to MongoDB
3. Runs continuously in the background

## Setup

Set the following environment variables:

```
RABBITMQ_URI=amqp://user:pass@rabbitmq:5672/
MONGO_URI=mongodb://mongo:27017
```

Run:
```bash
go run main.go
```

## Docker

```bash
docker build -t bestbuy-makeline-service .
docker run -p 8081:8081 \
  -e RABBITMQ_URI=amqp://user:pass@rabbitmq:5672/ \
  -e MONGO_URI=mongodb://mongo:27017 \
  bestbuy-makeline-service
```
