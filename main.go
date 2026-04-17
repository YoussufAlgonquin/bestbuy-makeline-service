package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/YoussufAlgonquin/bestbuy-makeline-service/db"
    "github.com/YoussufAlgonquin/bestbuy-makeline-service/queue"
)

func main() {
    // Connect to dependencies
    db.Connect()
    queue.StartConsuming()

    // Simple health server — Kubernetes needs this to know the pod is alive
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "makeline-service"})
    })

    port := "8081"
    log.Printf("makeline-service health server on :%s", port)
    r.Run(":" + port)
}