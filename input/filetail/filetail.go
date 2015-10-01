package filetail

import (
	"encoding/json"
	"github.com/DLag/logear/basiclogger"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const module = "filetail"

type FileTail struct {
	tag          string
	paths        []string
	files        map[string]*tail.Tail
	timestamp    string
	messageQueue chan *basiclogger.Message
}

//TODO: directory for .pos files
func Init(messageQueue chan *basiclogger.Message, conf map[string]interface{}) *FileTail {
	var paths []string
	for _, path := range conf["path"].([]interface{}) {
		paths = append(paths, path.(string))
	}
	v := &FileTail{tag: conf["tag"].(string), messageQueue: messageQueue, paths: paths, timestamp: conf["timestamp"].(string)}
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
	log.Printf("[%s] File watcher started", v.tag)
	for {
		for file, _ := range v.files {
			select {
			case <-v.files[file].Dying():
				log.Printf("[%s] File \"%s\" closed", v.tag, file)
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
					log.Printf("[%s] Found file \"%s\"", v.tag, file)
					tc := tail.Config{Follow: true, ReOpen: false, MustExist: true, Poll: true, Logger: tail.DiscardingLogger}
					if s, err := ioutil.ReadFile(file + ".pos"); err == nil {
						s := strings.Split(string(s), "\n")
						if len(s) == 2 {
							if ctime, err := strconv.Atoi(s[0]); err == nil && int64(ctime) == ctimeFile(file) {
								if pos, err := strconv.Atoi(s[1]); err == nil {
									tc.Location = &tail.SeekInfo{Whence: os.SEEK_CUR, Offset: int64(pos)}
									log.Printf("[%s] Restoring position %d in file \"%s\"", v.tag, pos, file)
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
		var j map[string]interface{}
		err := json.Unmarshal([]byte(data.Text), &j)
		if err == nil {
			j["file"] = filepath.Base(t.Filename)
			j["@timestamp"] = j[v.timestamp]
			v.messageQueue <- &basiclogger.Message{Time: time.Now(), Data: j}
		} else {
			log.Printf("[%s] Bogus message in \"%s\"", v.tag, t.Filename)
		}
		pos, _ := t.Tell()
		ctime := ctimeFile(t.Filename)
		posstr := strconv.Itoa(int(ctime)) + "\n" + strconv.Itoa(int(pos))
		ioutil.WriteFile(t.Filename+".pos", []byte(posstr), 0644)
	}
	log.Printf("[%s] Closing file \"%s\"", v.tag, t.Filename)
	t.Stop()
}
