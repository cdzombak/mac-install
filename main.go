package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cdzombak/mac-install/internal/config"
	"github.com/cdzombak/mac-install/internal/orchestrator"
)

var version = "<dev>"

func main() {
	var configFile string
	var skipOptional bool
	var onlyTarget string
	var versionFlag bool
	flag.StringVar(&configFile, "config", "./install.yaml", "Path to configuration YAML file")
	flag.BoolVar(&skipOptional, "skip-optional", false, "Skip all optional sections")
	flag.StringVar(&onlyTarget, "only", "", "Install only a single piece of software matching this name")
	flag.BoolVar(&versionFlag, "version", false, "Print version and exit")
	flag.Parse()

	if versionFlag {
		printVersion()
		os.Exit(0)
	}

	if runtime.GOOS != "darwin" {
		log.Fatal("This program is designed to run on macOS only")
	}

	if configFile == "" {
		log.Fatal("Configuration file not specified")
	}

	if skipOptional && onlyTarget != "" {
		log.Fatal("Cannot use -skip-optional and -only flags together")
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get the directory containing the config file
	configDir := filepath.Dir(configFile)
	absConfigDir, err := filepath.Abs(configDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path of config directory: %v", err)
	}

	orchestrator := orchestrator.New(cfg, absConfigDir)
	orchestrator.SetSkipOptional(skipOptional)
	orchestrator.SetOnlyTarget(onlyTarget)
	if err := orchestrator.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("mac-install version %s\n", version)
}
