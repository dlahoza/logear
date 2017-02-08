package main

import (
	"github.com/BurntSushi/toml"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/hashicorp/logutils"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"reflect"
	"strconv"
)

var (
	cfg config
	logFilter *logutils.LevelFilter
	logLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"}
)

const (
	logMinLevel = logutils.LogLevel("INFO")
	envDelimiter = "_"
)

type config struct {
	sectionName	string
	data		map[string]interface{}
}

func readConfig() {
	var (
		configFile            string
		showHelp, showVersion bool
	)
	logFilter = &logutils.LevelFilter{
		Levels:   logLevels,
		MinLevel: logMinLevel,
		Writer:   os.Stderr,
	}
	log.SetOutput(logFilter)
	flag.StringVar(&configFile, []string{"c", "-config"}, "/etc/logear/logear.conf", "config file")
	flag.StringVar(&logFile, []string{"l", "-log"}, "", "log file")
	flag.BoolVar(&showHelp, []string{"h", "-help"}, false, "display the help")
	flag.BoolVar(&showVersion, []string{"v", "-version"}, false, "display version info")
	flag.Parse()
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if showVersion {
		println(versionstring)
		println("OS: " + runtime.GOOS)
		println("Architecture: " + runtime.GOARCH)
		os.Exit(0)
	}
	cfg.parseTomlFile(configFile)
	startLogging()
	log.Printf("%s started with pid %d", versionstring, os.Getpid())
}

func (c *config) parseTomlFile(filename string) {
	if data, err := ioutil.ReadFile(filename); err != nil {
		log.Fatal("[ERROR] Can't read config file ", filename, ", error: ", err)
	} else {
		if _, err := toml.Decode(string(data), &c.data); err != nil {
			log.Fatal("[ERROR] Can't parse config file ", filename, ", error: ", err)
		}
	}
}

func (c *config) getSection(name string) (interface{}, bool) {
	data, ok := c.data[name];
	if !ok {
		log.Printf("[WARNING] Don't found section named: %s", name)
		return nil, false
	}
	data = overrideByEnv(data, strings.ToLower(progname + envDelimiter + name))

	return data, true
}

func overrideByEnv(originalData interface{}, sectionName string) interface{} {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, sectionName) {
			envNameValue := strings.SplitN(env, "=", 2)
			envNameValue[0] = strings.TrimPrefix(envNameValue[0], sectionName + envDelimiter)

			switch reflect.TypeOf(originalData).Kind()  {
			case reflect.Map:
				data := originalData.(map[string]interface{})
				applyInMap(envNameValue, data)
				originalData = data
			case reflect.Slice:
				data := originalData.([]map[string]interface{})
				items := strings.SplitN(envNameValue[0], envDelimiter, 2)
				if index, err := strconv.Atoi(items[0]); err == nil {
					envNameValue[0] = items[1]
					lastIndex := len(data) - 1
					if index > lastIndex {
						newData := make(map[string]interface{})
						applyInMap(envNameValue, newData)
						data = append(data, newData)
					} else {
						applyInMap(envNameValue, data[index])
					}
				}
				originalData = data
			}
		}
	}
	return originalData
}

func applyInMap(env []string, data map[string]interface{}) {
	var propName string
	var index int
	var isArray bool = false

	items := strings.Split(env[0], envDelimiter)
	if len(items) > 1 {
		if i, err := strconv.Atoi(items[len(items)-1]); err == nil {
			propName = strings.Join(items[:len(items)-1], envDelimiter)
			index = i
			isArray = true
		} else {
			propName = env[0]
		}
	} else {
		propName = env[0]
	}

	if _, ok := data[propName]; ok {
		switch reflect.TypeOf(data[propName]).Kind() {
		case reflect.String:
			data[propName] = env[1]
		case reflect.Slice:
			lastIndex := len(data[propName].([]interface{})) - 1
			if index > lastIndex {
				data[propName] = append(data[propName].([]interface{}), env[1])
			} else {
				data[propName].([]interface{})[index] = env[1]
			}
		}
	} else if isArray {
		data[propName] = append([]interface{}{}, env[1])
	} else {
		data[propName] = env[1]
	}

}
