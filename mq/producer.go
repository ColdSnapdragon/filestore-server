package mq

import (
	"filestore-server/config"
	"github.com/streadway/amqp"
	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel

// 检查Channel状态
func initChannel() bool {
	if channel != nil {
		return true
	}
	c, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println("连接rabbitmq失败:", err)
		return false
	}
	conn = c
	channel, err = conn.Channel()
	if err != nil {
		log.Println("建立channel失败:", err)
		return false
	}
	return true
}

// Publish 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	// 1.判断channel是否正常
	if !initChannel() {
		return false
	}

	err := channel.ExchangeDeclare(
		config.TransExchangeName, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		log.Println("声明交换机失败:", err)
		return false
	}

	// 2.执行消息发布
	err = channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		log.Println("发布消息失败:", err)
		return false
	}
	return true
}
