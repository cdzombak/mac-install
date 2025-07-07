package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/cdzombak/mac-install/internal/config"
	"github.com/cdzombak/mac-install/internal/orchestrator"
)

// TODO(cdzombak): exitcode
// TODO(cdzombak): improve UX for missing args etc here
// TODO(cdzombak): versioning
// TODO(cdzombak): standard build process

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration YAML file")
	flag.Parse()

	if runtime.GOOS != "darwin" {
		log.Fatal("This program is designed to run on macOS only")
	}

	if configFile == "" {
		log.Fatal("Configuration file not specified")
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	orchestrator := orchestrator.New(cfg)
	if err := orchestrator.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
		os.Exit(1)
	}
}
