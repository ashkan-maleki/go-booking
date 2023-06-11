package config

const (
	Exchange          = "webapp.reservation"
	QueueOrchestrator = "orchestrator.reservation"
	QueueValidation   = "validation"
	QueueReservation  = "reservation"
	QueueCredit       = "credit"
	QueueBooking      = "booking"
	RabbitMQUri       = "amqp://guest:guest@localhost:5672"
)
