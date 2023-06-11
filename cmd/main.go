package main

import (
	"github.com/mamalmaleki/go-advanced-samples/internal/broker/queue"
	"log"
)

func main() {
	log.Println("starting booking service")
	queue.Setup()
}
