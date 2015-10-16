package main

import (
	"github.com/BurntSushi/toml"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/hashicorp/logutils"
	"io/ioutil"
	"log"
	"os"
)

var (
	cfg       map[string]interface{}
	logFilter *logutils.LevelFilter
	logLevels = []logutils.LogLevel{"DEBUG", "WARN", "ERROR"}
)

const (
	logMinLevel = logutils.LogLevel("DEBUG")
)

func parseTomlFile(filename string) {
	if data, err := ioutil.ReadFile(filename); err != nil {
		log.Fatal("[ERROR] Can't read config file ", filename, ", error: ", err)
	} else {
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			log.Fatal("[ERROR] Can't parse config file ", filename, ", error: ", err)
		}
	}
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
		os.Exit(0)
	}
	parseTomlFile(configFile)
	startLogging()
	log.Printf("%s %s started with pid %d", progname, version, os.Getpid())
}
