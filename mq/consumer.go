package mq

// 封装一些操作逻辑，不涉及真正业务

import "log"

var done chan struct{}

// StartConsume 开始监听队列，获取消息
func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	if !initChannel() {
		return
	}

	_, err := channel.QueueDeclare(
		qName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("声明队列失败:", err)
		return
	}

	msgs, err := channel.Consume(
		qName,
		cName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("监听队列失败:", err)
		return
	}

	done = make(chan struct{})
	go func() {
		for msg := range msgs {
			ok := callback(msg.Body)
			if !ok {
				// TODO:写到另一个队列，用于异常情况重试
			}
		}
	}()

	<-done
	channel.Close()
}
