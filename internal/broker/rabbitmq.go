package broker

import (
	"github.com/mamalmaleki/go-advanced-samples/cmd/config"
	amqp "github.com/rabbitmq/amqp091-go"
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


