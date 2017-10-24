package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/zerok/tpl/internal/world"
)

var version, commit, date string

func main() {
	log := logrus.New()
	var input string
	var showVersion bool
	var verbose bool

	pflag.BoolVar(&verbose, "verbose", false, "Verbose log output")
	pflag.BoolVar(&showVersion, "version", false, "Show version information")
	pflag.Parse()

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	if showVersion {
		fmt.Printf("Version: %s\nCommit: %s\nBuild date: %s\n", version, commit, date)
		os.Exit(0)
	}

	input = pflag.Arg(0)
	if input == "" {
		log.Fatalf("No input file provided")
	}
	fp, err := os.Open(input)
	if err != nil {
		log.WithError(err).Fatalf("Failed to open template %s")
	}
	defer fp.Close()

	w := world.World{}
	if err := w.Render(os.Stdout, fp); err != nil {
		log.WithError(err).Fatal("Failed to render")
	}
}
