package main

import (
	basiclogger "./basiclogger"
)

func init() {
	readConfig()
}

func main() {
	basiclogger.InitMessageQueue(1)
	if outputs, ok := cfg["output"]; ok {
		outputs := outputs.([]map[string]interface{})
		for _, output := range outputs {
			basiclogger.AddOutput(InitOutput(output))
		}
	}
	if inputs, ok := cfg["input"]; ok {
		inputs := inputs.([]map[string]interface{})
		for _, input := range inputs {
			basiclogger.AddInput(InitInput(input))
		}
	}
	q := basiclogger.StartMessageQueue()
	<-q
}
