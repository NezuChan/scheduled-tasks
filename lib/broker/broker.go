package broker

import (
	"fmt"
	"github.com/disgoorg/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Broker struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	retryAttempt int
}

func CreateBroker(amqpURI string) *Broker {
	broker := &Broker{}
	var err error

	for {
		broker.Connection, err = amqp.DialConfig(amqpURI,
			amqp.Config{
				Dial: amqp.DefaultDial(time.Second * 15),
			})

		if err == nil {
			break
		}

		log.Fatalf("Failed to connect to RabbitMQ: %v, retrying in %d seconds", err, 10)
		time.Sleep(time.Second * 10)
	}

	broker.Channel, err = broker.Connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	go broker.handleReconnect(amqpURI)
	log.Infof("Connected to RabbitMQ server")
	return broker
}

func (broker *Broker) handleReconnect(amqpURI string) {
	onClose := broker.Connection.NotifyClose(make(chan *amqp.Error))
	for {
		<-onClose
		log.Warnf("RabbitMQ connection lost. Reconnecting...")
		for {
			if broker.retryAttempt >= 3 {
				panic(fmt.Sprintf("couldn't reconnect to amqp broker after %d attempts", broker.retryAttempt))
			}
			conn, err := amqp.DialConfig(amqpURI, amqp.Config{
				Dial: amqp.DefaultDial(time.Second * 5),
			})
			broker.retryAttempt += 1
			if err == nil {
				broker.Connection = conn
				broker.retryAttempt = 0
				break
			}
			log.Fatalf("Failed to connect to RabbitMQ: %v, retrying in %d seconds", err, 5)
			time.Sleep(time.Second * 5)
		}

		broker.Channel, _ = broker.Connection.Channel()
		broker.handleReconnect(amqpURI)
	}
}
