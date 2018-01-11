//go:generate license-notice
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
	var showLicenseInfo bool
	var leftDelim string
	var rightDelim string
	var insecure bool

	pflag.Usage = func() {
		fmt.Print("Usage: tpl [options] template-file\n\n")
		pflag.PrintDefaults()
	}

	pflag.StringVar(&vaultPrefix, "vault-prefix", "", "Prefix for all Vault paths")
	pflag.StringVar(&vaultMapping, "vault-mapping", "", "Key mapping file for Vault keys")
	pflag.BoolVar(&verbose, "verbose", false, "Verbose log output")
	pflag.BoolVar(&showVersion, "version", false, "Show version information")
	pflag.BoolVar(&showLicenseInfo, "licenses", false, "Show licenses of used libraries")
	pflag.StringVar(&leftDelim, "left-delimiter", "{{", "Left delimiter used within the Go template system")
	pflag.StringVar(&rightDelim, "right-delimiter", "}}", "Right delimiter used within the Go template system")
	pflag.BoolVar(&insecure, "insecure", false, "Enables features like shell output")
	pflag.Parse()

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	if showVersion {
		fmt.Printf("Version: %s\nCommit: %s\nBuild date: %s\n", version, commit, date)
		os.Exit(0)
	}

	if showLicenseInfo {
		fmt.Print("The following 3rd-party libraries have been used to create this project:\n\n\n\n")
		for _, li := range getLicenseInfos() {
			fmt.Printf("============================================================\n")
			fmt.Printf("https://%s\n", li.Package)
			fmt.Printf("------------------------------------------------------------\n\n")
			fmt.Printf("%s\n\n", li.LicenseText)
		}
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
		log.WithError(err).Fatalf("Failed to open template %s", input)
	}
	defer fp.Close()

	w := world.New(&world.Options{
		Logger:   log,
		Insecure: insecure,
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
