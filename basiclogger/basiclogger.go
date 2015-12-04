package basiclogger

import (
	"log"
	"time"
)

var (
	MessageQueue   chan *Message
	Outputs        = []Output{}
	Inputs         = []Input{}
	InputsCatalog  = map[string]InputInitFunc{}
	OutputsCatalog = map[string]OutputInitFunc{}
)

type InputInitFunc func(chan *Message, map[string]interface{}) Input
type OutputInitFunc func(map[string]interface{}) Output

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

func RegisterInput(name string, f InputInitFunc) {
	InputsCatalog[name] = f
}

func RegisterOutput(name string, f OutputInitFunc) {
	OutputsCatalog[name] = f
}

func InitMessageQueue(length int) {
	MessageQueue = make(chan *Message, length)
	for k, _ := range InputsCatalog {
		log.Printf("[DEBUG] Found input plugin: %s", k)
	}
	for k, _ := range OutputsCatalog {
		log.Printf("[DEBUG] Found output plugin: %s", k)
	}
}

func StartMessageQueue() chan bool {
	if len(Inputs) == 0 {
		log.Fatal("[ERROR] [messageQueue] Can't start without inputs. Please specify at least one.")
	}
	if len(Outputs) == 0 {
		log.Fatal("[ERROR] [messageQueue] Can't start without Outputs. Please specify at least one.")
	}
	for _, i := range Inputs {
		go i.Listener()
	}
	q := make(chan bool)
	go messageQueueWorker(q)
	return q
}

func AddOutput(o Output) {
	if o != nil {
		Outputs = append(Outputs, o)
		log.Printf("[INFO] [messageQueue] \"%s\" Output added to message queue", o.Tag())
	}
}

func AddInput(i Input) {
	if i != nil {
		Inputs = append(Inputs, i)
		log.Printf("[INFO] [messageQueue] \"%s\" Input added to message queue", i.Tag())
	}
}

func messageQueueWorker(q chan bool) {
	defer func() {
		q <- true
	}()
	var message *Message
	log.Print("[INFO] [messageQueue] Queue worker started")
	for {
		message = <-MessageQueue
		for _, v := range Outputs {
			if message != nil {
				v.Send(message)
				log.Printf("[DEBUG] [messageQueue] %v", message)
			}
		}
	}
}
