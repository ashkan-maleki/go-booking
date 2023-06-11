package queue

import (
	"github.com/mamalmaleki/go-advanced-samples/internal/broker"
	"log"
)

func Setup() {
	conn, connectionClose, err := broker.NewRabbitMQ()
	if err != nil {
		panic(err)
	}
	defer connectionClose()
	log.Println(conn)
}
