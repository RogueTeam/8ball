package main

import (
	"flag"
	"log"
	"os"

	"anarchy.ttfm/8ball/cmd/gateway/internal/router"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var app struct {
	debug  bool
	config string
}

func init() {
	flagset := flag.NewFlagSet("gatewat", flag.ExitOnError)
	flagset.BoolVar(&app.debug, "debug", false, "set debug mode")
	flagset.StringVar(&app.config, "config", "config.yaml", "YAML configuration")
	err := flagset.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if app.debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	configContents, err := os.ReadFile(app.config)
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	err = yaml.Unmarshal(configContents, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctrl, config, err := cfg.Compile()
	if err != nil {
		log.Fatal(err)
	}
	defer config.DB.Close()

	e := gin.Default()
	var r = router.Router{
		ProcessInterval: cfg.ProcessInterval,
		Gateway:         &ctrl,
		Base:            e,
	}
	r.Register()

	err = e.Run(cfg.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}
}
