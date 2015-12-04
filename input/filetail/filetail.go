package filetail

import (
	bl "github.com/DLag/logear/basiclogger"
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
	messageQueue     chan *bl.Message
}

func init() {
	bl.RegisterInput(module, Init)
}

//TODO: directory for .pos files
//TODO: config validation
func Init(messageQueue chan *bl.Message, conf map[string]interface{}) bl.Input {
	v := &FileTail{
		tag:              bl.GString("tag", conf),
		messageQueue:     messageQueue,
		paths:            bl.GArrString("path", conf),
		timestamp:        bl.GString("timestamp", conf),
		timestamp_format: bl.GString("timestamp_format", conf),
		filter:           bl.GString("filter", conf)}
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
	log.Printf("[INFO] [%s] File watcher started", v.tag)
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
		recvtime := time.Now()
		var m map[string]interface{}
		err := bl.FilterData(v.filter, data.Text, &m)
		log.Printf("[DEBUG] [%s] Recieved from filter: \"%v\"", v.tag, m)
		if err == nil {
			m["file"] = filepath.Base(t.Filename)
			if len(v.timestamp) > 0 {
				if timestamp, ok := m[v.timestamp]; ok {
					timestamp := timestamp.(string)
					if len(v.timestamp_format) > 0 {
						recvtime, err = time.Parse(v.timestamp_format, timestamp)
						if err != nil {
							log.Printf("[WARN] [%s] Bogus timestamp in \"%s\"", v.tag, t.Filename)
						}
					}
				}
			}
			v.messageQueue <- &bl.Message{Time: recvtime, Data: m}
		} else {
			log.Printf("[WARN] [%s] Bogus message in file \"%s\": %s", v.tag, t.Filename, data.Text)
		}
		pos, _ := t.Tell()
		ctime := ctimeFile(t.Filename)
		posstr := strconv.Itoa(int(ctime)) + "\n" + strconv.Itoa(int(pos))
		ioutil.WriteFile(t.Filename+".pos", []byte(posstr), 0644)
	}
	log.Printf("[DEBUG] [%s] Closing file \"%s\"", v.tag, t.Filename)
	t.Stop()
}
