package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

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
	_ = logger //пока не использую логгер

	sim := simulator.NewSimulator(&cfg.Simulator)

	run(sim)
}

func run(sim *simulator.Simulator) {
	fmt.Println("Choose mode: 1) Step 2) Auto")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		sim.PrintState()
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\nExit")
				break
			}

			input = strings.TrimSpace(input)

			if input == "q" {
				fmt.Println("End of simulation")
				break
			}

			if input == "" {
				if !sim.Step() {
					fmt.Println("End of simulation")
					break
				}
			} else {
				fmt.Println("Invalid input")
			}
		}

		sim.PrintFinalStats()

	case "2":
		sim.PrintState()
		for sim.Step() {
		}
		sim.PrintFinalStats()

	default:
		fmt.Println("Invalid input")
	}

}
