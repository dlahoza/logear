package main

import (
	"fmt"
	"time"
)

var (
	MessageQueue chan *Message
	Outputs      = map[string]Output{}
)

type Message struct {
	Time time.Time
	Data map[string]interface{}
}

type Output interface {
	//Init(conf map[string]interface{})
	Send(message *Message) error
}

type Input interface {
}

func StartMessageQueue(length int) {
	MessageQueue = make(chan *Message, length)
	go messageQueueWorker()
}

func messageQueueWorker() {
	var message *Message
	for {
		message = <-MessageQueue
		fmt.Println(message.Time, message.Data)
		for _, v := range Outputs {
			if message != nil {
				v.Send(message)
			}
		}
	}
}
