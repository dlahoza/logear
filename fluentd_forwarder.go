package main

import (
	"gopkg.in/vmihailenco/msgpack.v2"
	"log"
	"net"
	"time"
	//"encoding/json"
	"fmt"
	"os"
)

type FluentdforwarderOutput struct {
	c    chan *Message
	conn net.Conn
	host string
}

func FluentdforwarderOutputInit(conf ConfigFluentdForwarder) *FluentdforwarderOutput {
	return &FluentdforwarderOutput{
		c:    make(chan *Message),
		conn: nil,
		host: conf.Host}
}

func (v FluentdforwarderOutput) Send(message *Message) error {
	var err error
	now := time.Now().Unix()
	for {
		if v.conn == nil {
			v.conn, err = net.DialTimeout("tcp", v.host, time.Second*10)
			if err != nil {
				log.Printf("%v", err)
				if v.conn != nil {
					v.conn.Close()
				}
				time.Sleep(time.Second)
				continue
			}
		}
		message.Data["host"], _ = os.Hostname()
		val := []interface{}{"tinylog", now, message.Data}
		j, err := msgpack.Marshal(val)
		if err != nil {
			fmt.Printf("Bogus message: %v", message.Data)
			break
		}
		_, err = v.conn.Write(j)
		log.Printf("%s", string(j))
		if err != nil {
			log.Printf("%v", err)
			v.conn.Close()
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return err
}
