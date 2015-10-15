package main

import (
	"github.com/DLag/logear/basiclogger"
)

func init() {
	readConfig()
}

func main() {
	basiclogger.InitMessageQueue(1)
	//Initializing filters
	if filters, ok := cfg["filter"]; ok {
		filters := filters.([]map[string]interface{})
		for _, filter := range filters {
			basiclogger.AddFilter(filter)
		}
	}
	//Initializing outputs
	if outputs, ok := cfg["output"]; ok {
		outputs := outputs.([]map[string]interface{})
		for _, output := range outputs {
			basiclogger.AddOutput(InitOutput(output))
		}
	}
	//Initializing inputs
	if inputs, ok := cfg["input"]; ok {
		inputs := inputs.([]map[string]interface{})
		for _, input := range inputs {
			basiclogger.AddInput(InitInput(input))
		}
	}
	q := basiclogger.StartMessageQueue()
	<-q
}
