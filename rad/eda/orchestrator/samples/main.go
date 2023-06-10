package samples

import (
	"context"
	"fmt"
	"github.com/mamalmaleki/go-advanced-samples/rad/eda/orchestrator/cmd/config"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"strings"
)

// **************
// first code piece
// **************

var Exchange = "webapp.reservation"
var QueueOrchestrator = "orchestrator.reservation"
var QueueValidation = "validation"
var QueueReservation = "reservation"
var QueueCredit = "credit"
var QueueBooking = "booking"

// **************
// 2nd code piece
// **************

var exchange = config.Exchange
var queueNames = []string{
	config.QueueOrchestrator, config.QueueValidation, config.QueueReservation, config.QueueCredit, config.QueueValidation,
}

func main() {

	conn, closeConnection := mq.NewRabbitMQ()
	defer closeConnection()

	channel, err := conn.Channel()
	panicWithMessage(err, "couldn't create a channel")

	// Declare the primary exchange that binds to queues
	err = channel.ExchangeDeclare(exchange, "topic", true,
		false, false, false, nil)
	panicWithMessage(err, "couldn't declare validation exchange")

	for _, name := range queueNames {
		_, err = channel.QueueDeclare(name, true,
			false, false, false, nil)
		panicWithMessage(err, fmt.Sprintf("couldn't declare the %s queue", name))

		err = channel.QueueBind(name, name+".*", exchange, false, nil)
		panicWithMessage(err, fmt.Sprintf("couldn't bind the %s queue ", name))
	}

	log.Println("> declaration completed")

}

// **************
// 3rd code piece
// **************

func main() {

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	bookingRequestRepository := repository.NewRedis[types.BookingRequest](rdb)

	connection, closeConnection := mq.NewRabbitMQ()
	defer closeConnection()
	channel, err := connection.Channel()
	panicWithMessage(err, "error while establishing a channel")

	e := echo.New()
	e.GET("/book", func(c echo.Context) error {
		request := types.BookingRequest{
			UID:    shortid.MustGenerate(),
			UserID: c.QueryParam("user_id"),
			RoomID: c.QueryParam("room_id"),
		}
		err := channel.PublishWithContext(context.TODO(), config.Exchange, config.QueueOrchestrator+".new",
			false, false, amqp091.Publishing{
				Body: types.Encode(request),
			})
		if err != nil {
			return c.String(500, "something went wrong")
		}
		return c.String(201, fmt.Sprintf("request %s has been created!", request.UID))
	})

	e.GET("/book/:uid", func(c echo.Context) error {
		key := "bookingRequest:" + c.Param("uid")
		item, _ := bookingRequestRepository.Get(c.Request().Context(), key)
		return c.String(201, fmt.Sprintf("status=%s\ninfo=%s", item.State, item.Additional))
	})
	go func() {
		e.Logger.Fatal(e.Start(":1323"))
	}()

	mq.GracefullyExit()

}

// **************
// 4th code piece
// **************

type BookingRequest struct {
	UID    string `json:"UID"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`

	State      string `json:"state"`
	Additional string `json:"additional"`
}

// **************
// 5th code piece
// **************

func main() {

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	bookingRequestRepository := repository.NewRedis[types.BookingRequest](rdb)

	connection, closeConnection := mq.NewRabbitMQ()
	defer closeConnection()

	consumeChannel, err := connection.Channel()
	panicWithMessage(err, "couldn't establish a consume Channel")

	produceChannel, err := connection.Channel()
	panicWithMessage(err, "couldn't establish a produce Channel")

	queue, err := consumeChannel.QueueDeclare(config.QueueOrchestrator, true,
		false, false, false, nil)
	panicWithMessage(err, "couldn't declare queue")

	msgs, err := consumeChannel.Consume(queue.Name, "", true, false, false, false, nil)
	panicWithMessage(err, "error while creating the consumer")

	// updating the state of request and publishing to reservation queue
	handleNew := func(msg amqp091.Delivery) error {}

	handleValidated := func(msg amqp091.Delivery) error {}

	handleReserved := func(msg amqp091.Delivery) error {}

	handleDeposit := func(msg amqp091.Delivery) error {}

	handleBooked := func(msg amqp091.Delivery) error {}

	handleRefund := func(msg amqp091.Delivery) error {}

	actions := map[string]func(msg amqp091.Delivery) error{
		"new":       handleNew,
		"validated": handleValidated,
		"reserved":  handleReserved,
		"deposit":   handleDeposit,
		"booked":    handleBooked,
		"refund":    handleRefund,
	}

	go func() {
		// consuming messages
		for msg := range msgs {
			if err := mq.Handler(msg, config.QueueOrchestrator, actions); err != nil {
				// we should utilize deal-letter queues here, but I'll cover them in another article
				log.Printf("key:%s\tunhandled error: %s\n", msg.RoutingKey, err.Error())
			}
		}
	}()
	mq.GracefullyExit()
}

// **************
// 6th code piece
// **************

handleNew := func(msg amqp091.Delivery) error {
	request := types.Decode[types.BookingRequest](msg.Body)
	log.Printf("recieved a new booking request #%s\n", request.UID)
	request.State = "proceeding"
	bookingRequestRepository.Save(context.TODO(), request)

	return produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueValidation+".validate", false,
		false, amqp091.Publishing{
			Body: types.Encode(types.ValidationRequest{
				UID:    request.UID,
				UserID: request.UserID,
				RoomID: request.RoomID,
			}),
			ReplyTo: config.QueueOrchestrator + ".validated",
		})
}

// **************
// 7th code piece
// **************

handleValidation := func(msg amqp091.Delivery) error {
	validationRequest := types.Decode[types.ValidationRequest](msg.Body)
	validationRequest.Validated = true
	if validationRequest.UserID == "some_guy" && validationRequest.RoomID == "some_hotel" {
		validationRequest.Validated = false
		validationRequest.Errors = append(validationRequest.Errors, "you have been blocked by this hotel")
	}
	return produceChannel.PublishWithContext(context.TODO(), msg.Exchange, msg.ReplyTo, false, false,
		amqp091.Publishing{
			Body: types.Encode(validationRequest),
		})
}

actions := map[string]func(msg amqp091.Delivery) error{
	"validate": handleValidation,
}

// **************
// 8th code piece
// **************

handleReserve := func(msg amqp091.Delivery) error {
	request := types.Decode[types.ReservationRequest](msg.Body)
	// using a global lock
	request.Reserved = true
	if request.RoomID == "reserved" {
		request.Reserved = false
		request.Errors = append(request.Errors, "the room is already reserved by someone else")
	}
	return produceChannel.PublishWithContext(context.TODO(), msg.Exchange, msg.ReplyTo,
		false, false, amqp091.Publishing{
			Body: types.Encode(request),
		})
}

handleCancel := func(msg amqp091.Delivery) error {
	request := types.Decode[types.ReservationRequest](msg.Body)
	return rdb.Del(context.TODO(), "room:"+request.RoomID).Err()
}


// **************
// 9th code piece
// **************

handleDeposit := func(msg amqp091.Delivery) error {
	deposit := types.Decode[types.DepositRequest](msg.Body)
	deposit.Done = true
	if deposit.UserID == "poor_guy" {
		deposit.Done = false
		deposit.Errors = append(deposit.Errors, "insufficient credit")
	}
	return produceChannel.PublishWithContext(context.TODO(), msg.Exchange, msg.ReplyTo, false, false,
		amqp091.Publishing{
			Body: types.Encode(deposit),
		})
}

handleRefund := func(msg amqp091.Delivery) error {
	deposit := types.Decode[types.DepositRequest](msg.Body)
	// some business logic ...
	deposit.Done = true
	if deposit.UserID == "very_unlucky_guy" {
		deposit.Done = false
		deposit.Errors = append(deposit.Errors, "unexpected error")
	}
	return produceChannel.PublishWithContext(context.TODO(), msg.Exchange, msg.ReplyTo, false, false,
		amqp091.Publishing{
			Body: types.Encode(deposit),
		})
}

// **************
// 10th code piece
// **************

handleBooking := func(msg amqp091.Delivery) error {
	book := types.Decode[types.BookRequest](msg.Body)
	book.Done = true
	if book.UserID == "unlucky_guy" || book.UserID == "very_unlucky_guy" {
		book.Done = false
		book.Errors = append(book.Errors, "something unexpected happened")
	}
	return produceChannel.PublishWithContext(context.TODO(), msg.Exchange, msg.ReplyTo, false, false,
		amqp091.Publishing{
			Body: types.Encode(book),
		})
}

// **************
// 11th code piece
// **************

handleValidated := func(msg amqp091.Delivery) error {
	validation := types.Decode[types.ValidationRequest](msg.Body)
	request, _ := bookingRequestRepository.Get(context.TODO(), "bookingRequest:"+validation.UID)
	log.Printf("validation #%s completed: %t", validation.UID, validation.Validated)
	if validation.Validated {
		return produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueReservation+".reserve",
			false, false, amqp091.Publishing{
				ReplyTo: config.QueueOrchestrator + ".reserved",
				Body: types.Encode(types.ReservationRequest{
					UID:    request.UID,
					RoomID: request.RoomID,
					UserID: request.UserID,
				}),
			})
	} else {
		request.Additional = strings.Join(validation.Errors, ", ")
		request.State = "finished"
		return bookingRequestRepository.Save(context.TODO(), request)
	}
}

handleReserved := func(msg amqp091.Delivery) error {
	reserved := types.Decode[types.ReservationRequest](msg.Body)
	request, _ := bookingRequestRepository.Get(context.TODO(), "bookingRequest:"+reserved.UID)
	log.Printf("reservation #%s completed: %t", reserved.UID, reserved.Reserved)
	if reserved.Reserved {
		return produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueCredit+".deposit",
			false, false, amqp091.Publishing{
				ReplyTo: config.QueueOrchestrator + ".deposit",
				Body: types.Encode(types.DepositRequest{
					UID:    request.UID,
					UserID: request.UserID,
					Delta:  -2500, // in production, we need another service for this purpose
				}),
			})
	} else {
		request.Additional = strings.Join(reserved.Errors, ", ")
		request.State = "finished"
		return bookingRequestRepository.Save(context.TODO(), request)
	}
}

handleDeposit := func(msg amqp091.Delivery) error {
	deposit := types.Decode[types.DepositRequest](msg.Body)
	request, _ := bookingRequestRepository.Get(context.TODO(), "bookingRequest:"+deposit.UID)
	log.Printf("deposit #%s completed: %t", deposit.UID, deposit.Done)
	if deposit.Done {
		return produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueBooking+".book",
			false, false, amqp091.Publishing{
				ReplyTo: config.QueueOrchestrator + ".booked",
				Body: types.Encode(types.BookRequest{
					UID:    request.UID,
					UserID: request.UserID,
					RoomID: request.RoomID,
				}),
			})
	} else {
		request.Additional = strings.Join(deposit.Errors, ", ")
		request.State = "finished"
		return bookingRequestRepository.Save(context.TODO(), request)
	}
}

handleBooked := func(msg amqp091.Delivery) error {
	book := types.Decode[types.BookRequest](msg.Body)
	request, _ := bookingRequestRepository.Get(context.TODO(), "bookingRequest:"+book.UID)
	log.Printf("book #%s completed: %t", book.UID, book.Done)
	if book.Done {
		request.Additional = "booked successfully"
		request.State = "booked"
	} else {
		request.Additional = strings.Join(book.Errors, ", ")
		request.State = "finished"
		if err := produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueCredit+".refund",
			false, false, amqp091.Publishing{
				ReplyTo: config.QueueOrchestrator + ".refund",
				Body:    types.Encode(types.DepositRequest{UID: request.UID, Delta: 2500, UserID: request.UserID}),
			}); err != nil {
			return err
		}
	}
	if err := bookingRequestRepository.Save(context.TODO(), request); err != nil {
		return err
	}
	return produceChannel.PublishWithContext(context.TODO(), config.Exchange, config.QueueReservation+".cancel",
		false, false, amqp091.Publishing{
			Body: types.Encode(types.ReservationRequest{UID: request.UID}),
		})

}

handleRefund := func(msg amqp091.Delivery) error {
	deposit := types.Decode[types.DepositRequest](msg.Body)
	request, _ := bookingRequestRepository.Get(context.TODO(), "bookingRequest:"+deposit.UID)
	log.Printf("refund #%s completed: %t", deposit.UID, deposit.Done)
	if deposit.Done {
		request.Additional += ", refund completed."
		request.State = "refunded"
	} else {
		request.Additional += ", problem while refunding."
		request.State = "finished"
	}
	return bookingRequestRepository.Save(context.TODO(), request)
}

// **************
// 12th code piece
// **************