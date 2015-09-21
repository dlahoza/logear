package main

import (
	"fmt"
	"time"
)

const (
	progname = "TinyLog"
	version  = "0.0.1"
)

func init() {
	readConfig()
}

func main() {
	fmt.Printf("%v", cfg.FluentdForwarder.Host)
	Outputs["elastic"] = FluentdforwarderOutputInit(cfg.FluentdForwarder)
	StartMessageQueue(1)
	//MessageQueue <- &Message{Time: time.Now(), Data: map[string]interface{}{"message": "111"}}
	JsonFileInit(MessageQueue, cfg.Jsonfile)
	for {
		time.Sleep(time.Second)
	}
}
