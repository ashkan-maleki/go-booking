package broker

import (
	"github.com/mamalmaleki/go-advanced-samples/rad/eda/orchestrator/cmd/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

var (
	exchange   = config.Exchange
	queueNames = []string{
		config.QueueOrchestrator,
		config.QueueValidation,
		config.QueueReservation,
		config.QueueCredit,
		config.QueueValidation,
	}
)

type ConnectionClose func() error

func NewRabbitMQ() (*amqp.Connection, ConnectionClose, error) {
	conn, err := amqp.Dial(config.RabbitMQUri)
	if err != nil {
		return nil, nil, err
	}
	return conn, conn.Close, nil
}

func Setup() {
	conn, connectionClose, err := NewRabbitMQ()
	if err != nil {
		panic(err)
	}
	defer connectionClose()
	log.Println(conn)
}
