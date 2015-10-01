package main

import (
	"github.com/DLag/logear/basiclogger"
	"github.com/DLag/logear/input/filetail"
	"github.com/DLag/logear/output/fluentd_forwarder"
	"log"
)

func InitInput(input map[string]interface{}) basiclogger.Input {
	if v, ok := input["type"]; ok {
		t := v.(string)
		switch t {
		case "filetail":
			return filetail.Init(basiclogger.MessageQueue, input)
		default:
			log.Fatalf("\"%s\" isn't right input type", t)
		}
	} else {
		log.Fatal("You must specify type of input")
	}
	return nil
}

func InitOutput(output map[string]interface{}) basiclogger.Output {
	if v, ok := output["type"]; ok {
		t := v.(string)
		switch t {
		case "fluentd_forwarder":
			return fluentd_forwarder.Init(output)
		default:
			log.Fatalf("\"%s\" isn't right output type", t)
		}
	} else {
		log.Fatal("You must specify type of output")
	}
	return nil
}
