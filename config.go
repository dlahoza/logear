package main

import (
	"github.com/BurntSushi/toml"
	flag "github.com/docker/docker/pkg/mflag"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Elasticsearch    ConfigElasticsearch
	FluentdForwarder ConfigFluentdForwarder
	Jsonfile         ConfigJsonfile
}

type ConfigElasticsearch struct {
	Host  string
	Index string
}

type ConfigFluentdForwarder struct {
	Host string
}

type ConfigJsonfile struct {
	Path      []string
	Timestamp string
}

var cfg Config

func ParseTomlFile(filename string) error {
	if data, err := ioutil.ReadFile(filename); err != nil {
		log.Print("Can't read config file ", filename, ", error: ", err)
		return err
	} else {
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			log.Print("Can't parse config file ", filename, ", error: ", err)
			return err
		}
	}
	log.Print("Readed config file: ", filename)
	return nil
}

func ReadEnv(name string, val *string) {
	if s := os.Getenv(name); s != "" {
		*val = s
	}
}

func readConfig() {
	var (
		configFile string
		showHelp   bool
	)
	/*config.SetEnvPrefix("TINYLOG_")
	config.SetConfigName("tinylog")
	config.AddConfigPath("/etc/tinylog/")
	config.AddConfigPath("/usr/local/etc/tinylog/")
	config.AddConfigPath("./")*/

	flag.StringVar(&configFile, []string{"c", "-config"}, "/etc/tinylog/tinylog.toml", "config file")
	flag.BoolVar(&showHelp, []string{"h", "#help", "-help"}, false, "display the help")
	flag.Parse()
	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	ParseTomlFile(configFile)
	ReadEnv("ELASTICSEARCH_HOST", &cfg.Elasticsearch.Host)

}
