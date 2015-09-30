package fluentd_forwarder

import (
	basiclogger "../../basiclogger"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const module = "fluentd_forwarder"

type Fluentd_forwarder struct {
	tag     string
	c       chan *basiclogger.Message
	conn    net.Conn
	host    string
	timeout time.Duration
}

func Init(conf map[string]interface{}) *Fluentd_forwarder {
	if host, ok := conf["host"]; !ok {
		log.Fatal("[", module, "] You must specify host")
	} else {
		if port, ok := conf["port"]; !ok || port.(int64) <= 0 {
			log.Fatal("[", module, "] You must specify right port")
		} else {
			if timeout, ok := conf["timeout"]; !ok || timeout.(int64) <= 0 {
				log.Fatal("[", module, "] You must specify right timeout")
			} else {
				tag, _ := conf["tag"]
				return &Fluentd_forwarder{
					tag:     tag.(string),
					c:       make(chan *basiclogger.Message),
					conn:    nil,
					host:    host.(string) + ":" + strconv.Itoa(int(port.(int64))),
					timeout: time.Second * time.Duration(timeout.(int64))}
			}
		}
	}
	return nil
}

func (v Fluentd_forwarder) Tag() string {
	return v.tag
}

//TODO: SSL support
func (v Fluentd_forwarder) Send(message *basiclogger.Message) error {
	var err error
	now := time.Now().Unix()
	for {
		if v.conn == nil {
			if v.conn, err = net.DialTimeout("tcp", v.host, v.timeout); err != nil {
				log.Printf("%v", err)
				if v.conn != nil {
					v.conn.Close()
				}
				time.Sleep(time.Second)
				continue
			}
		}
		message.Data["host"], _ = os.Hostname()
		val := []interface{}{v.tag, now, message.Data}
		m, err := msgpack.Marshal(val)
		if err != nil {
			fmt.Printf("[%s] Bogus message: %v", v.tag, message.Data)
			break
		}
		v.conn.SetDeadline(time.Now().Add(v.timeout))
		_, err = v.conn.Write(m)
		if err != nil {
			log.Printf("[%s] Socket write error %v", v.tag, err)
			v.conn.Close()
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return err
}
