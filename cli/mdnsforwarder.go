package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"encoding/json"

	mdnsforwarder "github.com/cbrand/mdnsforwarder"
	log "github.com/sirupsen/logrus"

	cli "github.com/urfave/cli/v2"
)

func convertLogLevel(logLevel string) log.Level {
	switch strings.ToLower(logLevel) {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.WarnLevel
	}
}

func initLogLevel(logLevel string) {
	log.SetLevel(convertLogLevel(logLevel))
}

type Config struct {
	MdnsInterfaces []string `json:"mdnsInterfaces"`
	Listeners      []string `json:"listeners"`
	Targets        []string `json:"targets"`
}

func (config *Config) ToForwarder() (mdnsforwarder.Forwarder, error) {
	convertedInterfaces := []*net.Interface{}
	for _, mdnsInterfaceString := range config.MdnsInterfaces {
		networkInterface, err := net.InterfaceByName(mdnsInterfaceString)
		if err != nil {
			return nil, err
		}
		convertedInterfaces = append(convertedInterfaces, networkInterface)
	}

	convertedListeners := []*net.UDPAddr{}
	for _, listenerString := range config.Listeners {
		listener, err := net.ResolveUDPAddr("udp", listenerString)
		if err != nil {
			return nil, err
		}
		convertedListeners = append(convertedListeners, listener)
	}

	convertedTargets := []*net.UDPAddr{}
	for _, targetString := range config.Targets {
		target, err := net.ResolveUDPAddr("udp", targetString)
		if err != nil {
			return nil, err
		}
		convertedTargets = append(convertedTargets, target)
	}

	forwarderConfig := mdnsforwarder.New(convertedInterfaces, convertedListeners, convertedTargets)
	return forwarderConfig, nil
}

var app = &cli.App{
	Name:    "mdnsforwarder",
	Usage:   "Handler to forward mdns traffic between networks and to other forwarder instances",
	Version: "1.0.0",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"level"},
			EnvVars: []string{"LOG_LEVEL"},
			Value:   "warn",
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Configuration flie where the forwarder is configured",
			Value:   "/etc/config/mdnsforwarder",
		},
	},
	Action: func(c *cli.Context) error {
		initLogLevel(c.String("log-level"))

		configPath := c.String("config")
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				log.Fatal("Couldn't read raw data from config file")
				return err
			}
			config := &Config{}
			err = json.Unmarshal(data, config)
			if err != nil {
				log.Fatal("Config file doesn't have valid json construct")
				return err
			}
			forwarder, err := config.ToForwarder()
			if err != nil {
				log.Fatal(fmt.Sprintf("Failed to convert config file: %s", err.Error()))
				return err
			}
			return forwarder.Run()
		} else if os.IsNotExist(err) {
			log.Fatal(fmt.Sprintf("Config file not found at %s", configPath))
			return err
		} else {
			log.Fatal("Unknown error when getting file stats")
			return err
		}
	},
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
