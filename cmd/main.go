package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/config"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/logger"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/simulator"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "config path")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config parse fail: %s\n", err)
		os.Exit(1)
	}

	logger, err := logger.Setup(&cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup logger fail: %s\n", err)
		os.Exit(1)
	}

	sim := simulator.NewSimulator(&cfg.Simulator)

	_ = sim
	_ = logger
}
