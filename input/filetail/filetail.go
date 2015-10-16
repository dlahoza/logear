package filetail

import (
	"github.com/DLag/logear/basiclogger"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const module = "filetail"

type FileTail struct {
	tag              string
	paths            []string
	files            map[string]*tail.Tail
	timestamp        string
	timestamp_format string
	filter           string
	messageQueue     chan *basiclogger.Message
}

//TODO: directory for .pos files
//TODO: config validation
func Init(messageQueue chan *basiclogger.Message, conf map[string]interface{}) *FileTail {
	var paths []string
	for _, path := range conf["path"].([]interface{}) {
		paths = append(paths, path.(string))
	}
	v := &FileTail{
		tag:              conf["tag"].(string),
		messageQueue:     messageQueue,
		paths:            paths,
		timestamp:        conf["timestamp"].(string),
		timestamp_format: conf["timestamp_format"].(string),
		filter:           conf["filter"].(string)}
	return v
}

func (v *FileTail) Tag() string {
	return v.tag
}

func (v *FileTail) Listener() {
	v.files = make(map[string]*tail.Tail)
	go v.watcher()
}

func (v *FileTail) watcher() {
	log.Printf("[DEBUG] [%s] File watcher started", v.tag)
	for {
		for file, _ := range v.files {
			select {
			case <-v.files[file].Dying():
				log.Printf("[DEBUG] [%s] File \"%s\" closed", v.tag, file)
				delete(v.files, file)
			default:
				continue
			}
		}
		//Search for new files
		for _, path := range v.paths {
			f, _ := filepath.Glob(path)
			for _, file := range f {
				if _, ok := v.files[file]; !ok {
					log.Printf("[DEBUG] [%s] Found file \"%s\"", v.tag, file)
					tc := tail.Config{Follow: true, ReOpen: false, MustExist: true, Poll: true, Logger: tail.DiscardingLogger}
					if s, err := ioutil.ReadFile(file + ".pos"); err == nil {
						s := strings.Split(string(s), "\n")
						if len(s) == 2 {
							if ctime, err := strconv.Atoi(s[0]); err == nil && int64(ctime) == ctimeFile(file) {
								if pos, err := strconv.Atoi(s[1]); err == nil {
									tc.Location = &tail.SeekInfo{Whence: os.SEEK_CUR, Offset: int64(pos)}
									log.Printf("[DEBUG] [%s] Restoring position %d in file \"%s\"", v.tag, pos, file)
								}
							}
						}
					}
					t, err := tail.TailFile(file, tc)
					if err == nil {
						go v.worker(t)
						v.files[file] = t
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func (v *FileTail) worker(t *tail.Tail) {
	for data := range t.Lines {
		var m map[string]interface{}
		err := basiclogger.FilterData(v.filter, data.Text, &m)
		log.Printf("[DEBUG] [%s] Recieved from filter: \"%v\"", v.tag, m)
		if err == nil {
			m["file"] = filepath.Base(t.Filename)
			if len(v.timestamp) > 0 {
				if timestamp, ok := m[v.timestamp]; ok {
					timestamp := timestamp.(string)
					if len(v.timestamp_format) > 0 {
						timestamp = basiclogger.ConvertTimestamp(v.timestamp_format, timestamp)
						if len(timestamp) == 0 {
							log.Printf("[WARN] [%s] Bogus timestamp in \"%s\"", v.tag, t.Filename)

						} else {
							m["@timestamp"] = timestamp
						}
					}
				}
			}
			v.messageQueue <- &basiclogger.Message{Time: time.Now(), Data: m}
		} else {
			log.Printf("[WARN] [%s] Bogus message in \"%s\"", v.tag, t.Filename)
		}
		pos, _ := t.Tell()
		ctime := ctimeFile(t.Filename)
		posstr := strconv.Itoa(int(ctime)) + "\n" + strconv.Itoa(int(pos))
		ioutil.WriteFile(t.Filename+".pos", []byte(posstr), 0644)
	}
	log.Printf("[DEBUG] [%s] Closing file \"%s\"", v.tag, t.Filename)
	t.Stop()
}
