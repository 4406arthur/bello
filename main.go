package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	core "github.com/4406arthur/bello/cmd"
	"github.com/4406arthur/bello/utils/logger"

	"github.com/spf13/viper"
)

var usageStr = `
AI Customer service controller

Server Options:
    -c, --config <file>              Configuration file path
    -h, --help                       Show this message
    -v, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func setup(path string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName("config")
	if path != "" {
		v.AddConfigPath(path)
	} else {
		v.AddConfigPath("./config/")
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	return v
}

var version string

func printVersion() {

	fmt.Printf(`AI Customer service controller %s, Compiler: %s %s, Copyright (C) 2020 EsunBank, Inc.`,
		version,
		runtime.Compiler,
		runtime.Version())
	fmt.Println()
}

func main() {
	//from env
	var configFile string
	var showVersion bool
	version = "0.0.1"
	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	flag.StringVar(&configFile, "c", "", "Configuration file path.")
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}
	viperConfig := setup(configFile)
	log := logger.InitLogger(
		viperConfig.GetString("server_config.log_path"),
		viperConfig.GetString("server_config.elasticsearch_endpoint"),
	)

	core.InitRouter(log, viperConfig)
}
