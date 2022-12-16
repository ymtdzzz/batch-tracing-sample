package internal

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPCarrier struct {
	headers amqp.Table
}

func (c *AMQPCarrier) Get(key string) string {
	return fmt.Sprintf("%s", c.headers[key])
}

func (c *AMQPCarrier) Set(key string, value string) {
	c.headers[key] = value
}

func (c *AMQPCarrier) Keys() []string {
	keys := make([]string, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}

func NewAMQPCarrier(headers amqp.Table) *AMQPCarrier {
	return &AMQPCarrier{headers}
}
