package basiclogger

import (
	"log"
	"time"
)

var (
	MessageQueue chan *Message
	Outputs      = []Output{}
	Inputs       = []Input{}
)

type Message struct {
	Time time.Time
	Data map[string]interface{}
}

type Output interface {
	Tag() string
	Send(message *Message) error
}

type Input interface {
	Listener()
	Tag() string
}

func InitMessageQueue(length int) {
	MessageQueue = make(chan *Message, length)

}

func StartMessageQueue() {
	if len(Inputs) == 0 {
		log.Fatal("[messageQueue] Can't start without inputs. Please specify at least one.")
	}
	if len(Outputs) == 0 {
		log.Fatal("[messageQueue] Can't start without Outputs. Please specify at least one.")
	}
	for _, i := range Inputs {
		go i.Listener()
	}
	go messageQueueWorker()
}

func AddOutput(o Output) {
	if o != nil {
		Outputs = append(Outputs, o)
		log.Printf("[messageQueue] \"%s\" Output added to message queue", o.Tag())
	}
}

func AddInput(i Input) {
	if i != nil {
		Inputs = append(Inputs, i)
		log.Printf("[messageQueue] \"%s\" input added to message queue", i.Tag())
	}
}

func messageQueueWorker() {
	var message *Message
	log.Print("[messageQueue] Queue worker started")
	for {
		message = <-MessageQueue
		for _, v := range Outputs {
			if message != nil {
				v.Send(message)
				log.Printf("[messageQueue] %v", message)
			}
		}
	}
}
