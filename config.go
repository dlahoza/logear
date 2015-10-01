package main

import (
	"github.com/BurntSushi/toml"
	flag "github.com/docker/docker/pkg/mflag"
	"io/ioutil"
	"log"
	"os"
)

var cfg map[string]interface{}

func parseTomlFile(filename string) {
	if data, err := ioutil.ReadFile(filename); err != nil {
		log.Fatal("Can't read config file ", filename, ", error: ", err)
	} else {
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			log.Fatal("Can't parse config file ", filename, ", error: ", err)
		}
	}
}

func initLogger(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file ", filename, ", error: ", err)
	}
	log.SetOutput(file)
}

func readConfig() {
	var (
		configFile string
		logFile    string
		showHelp   bool
	)

	flag.StringVar(&configFile, []string{"c", "-config"}, "/etc/logear/logear.conf", "config file")
	flag.StringVar(&logFile, []string{"l", "-log"}, "", "log file")
	flag.BoolVar(&showHelp, []string{"h", "-help"}, false, "display the help")
	flag.Parse()
	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	parseTomlFile(configFile)
	if logFile != "" {
		initLogger(logFile)
	} else {
		if v, ok := cfg["main"]; ok {
			if v, ok := v.(map[string]interface{})["logfile"]; ok {
				initLogger(v.(string))
			}
		}
	}
	log.Printf("%s %s started", progname, version)
}
