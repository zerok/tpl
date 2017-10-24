package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/zerok/tpl/internal/world"
)

var version, commit, date string

func main() {
	log := logrus.New()
	var input string
	var vaultPrefix string
	var showVersion bool
	var vaultMapping string
	var verbose bool

	pflag.Usage = func() {
		fmt.Println("Usage: tpl [options] template-file\n")
		pflag.PrintDefaults()
	}

	pflag.StringVar(&vaultPrefix, "vault-prefix", "", "Prefix for all Vault paths")
	pflag.StringVar(&vaultMapping, "vault-mapping", "", "Key mapping file for Vault keys")
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
		log.Error("No input file provided")
		pflag.Usage()
		os.Exit(1)
	}
	fp, err := os.Open(input)
	if err != nil {
		log.WithError(err).Fatalf("Failed to open template %s")
	}
	defer fp.Close()

	w := world.New(&world.Options{
		Logger: log,
	})
	if vaultPrefix != "" {
		w.Vault().Prefix = vaultPrefix
	}
	if vaultMapping != "" {
		mapping, err := loadKeyMapping(vaultMapping)
		if err != nil {
			log.WithError(err).Fatalf("Failed to load mapping file")
		}
		w.Vault().KeyMapping = mapping
	}
	if err := w.Render(os.Stdout, fp); err != nil {
		log.WithError(err).Fatal("Failed to render")
	}
}

func loadKeyMapping(path string) (map[string]string, error) {
	result := make(map[string]string)
	if path == "" {
		return result, nil
	}
	fp, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", path)
	}
	defer fp.Close()
	reader := csv.NewReader(fp)
	reader.Comma = ';'
	for {
		records, err := reader.Read()
		if err != nil {
			if io.EOF == err {
				break
			}
			return nil, errors.Wrap(err, "failed to read record from mapping file")
		}
		if len(records) != 2 {
			continue
		}
		result[records[0]] = records[1]
	}
	return result, nil
}
