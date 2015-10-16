package main

import (
	"github.com/hashicorp/logutils"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	reloadLogs chan os.Signal
	logFile    string
)

func openFileLog(filename string) io.Writer {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("[ERROR] Failed to open log file ", filename, ", error: ", err)
	}
	return file
}

func startLogging() {
	logOpen()
	reloadLogs = make(chan os.Signal)
	signal.Notify(reloadLogs, os.Interrupt, os.Kill, syscall.SIGUSR1)
	go logWatcher()
}

func logOpen() {
	if logFile != "" {
		logFilter.Writer = openFileLog(logFile)
	} else {
		if v, ok := cfg["main"]; ok {
			if v, ok := v.(map[string]interface{})["logfile"]; ok {
				logFilter.Writer = openFileLog(v.(string))
			}
			if v, ok := v.(map[string]interface{})["loglevel"]; ok {
				logFilter.MinLevel = logutils.LogLevel(v.(string))
			}
		}
	}
}

func logWatcher() {
	for {
		sig := <-reloadLogs
		if sig == syscall.SIGUSR1 {
			logOpen()
			log.Print("[DEBUG] Logfile reopen succesful")
		}
		if sig == os.Interrupt || sig == os.Kill {
			return
		}
	}
}
