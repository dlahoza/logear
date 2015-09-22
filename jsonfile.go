package main

import (
	"encoding/json"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

type JsonFile struct {
	files        []string
	timestamp    string
	messageQueue chan *Message
}

func JsonFileInit(messageQueue chan *Message, conf ConfigJsonfile) *JsonFile {
	var files []string
	for _, path := range conf.Path {
		f, _ := filepath.Glob(path)
		files = append(files, f...)
	}
	v := &JsonFile{messageQueue: messageQueue, files: files, timestamp: conf.Timestamp}
	log.Printf("%v", files)
	go v.Listener()
	return v
}

func (v *JsonFile) Listener() {
	for _, file := range v.files {
		t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true})
		log.Print(file)
		if err == nil {
			go v.worker(t)
		}
	}
}

func (v *JsonFile) worker(t *tail.Tail) {
	for data := range t.Lines {
		var j map[string]interface{}
		err := json.Unmarshal([]byte(data.Text), &j)
		if err == nil {
			j["file"] = filepath.Base(t.Filename)
			j["@timestamp"] = j[v.timestamp]
			v.messageQueue <- &Message{Time: time.Now(), Data: j}
			ioutil.WriteFile(t.Filename+".pos", strconv.Itoa(t.Tell), 0644)
		}
		log.Println(data.Text)
	}
}
