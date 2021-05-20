//go:generate license-notice
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	var data []string
	var azurePrefix string
	var azureMapping string
	var outputFile string

	pflag.Usage = func() {
		fmt.Print("Usage: tpl [options] template-file\n\n")
		pflag.PrintDefaults()
	}

	pflag.StringVar(&outputFile, "output", "", "Output file")
	pflag.StringVar(&vaultPrefix, "vault-prefix", "", "Prefix for all Vault paths")
	pflag.StringVar(&vaultMapping, "vault-mapping", "", "Key mapping file for Vault keys")
	pflag.BoolVar(&verbose, "verbose", false, "Verbose log output")
	pflag.BoolVar(&showVersion, "version", false, "Show version information")
	pflag.BoolVar(&showLicenseInfo, "licenses", false, "Show licenses of used libraries")
	pflag.StringVar(&leftDelim, "left-delimiter", "{{", "Left delimiter used within the Go template system")
	pflag.StringVar(&rightDelim, "right-delimiter", "}}", "Right delimiter used within the Go template system")
	pflag.BoolVar(&insecure, "insecure", false, "Enables features like shell output")
	pflag.StringSliceVar(&data, "data", []string{}, "Data definitions (e.g. --data=name=file.yaml)")
	pflag.StringVar(&azurePrefix, "azure-prefix", "", "Prefix for all Azure keyvault paths")
	pflag.StringVar(&azureMapping, "azure-mapping", "", "Key mapping file for Azure keyvault keys")
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

	var rd io.Reader
	if input == "-" {
		rd = os.Stdin
	} else {
		fp, err := os.Open(input)
		if err != nil {
			log.WithError(err).Fatalf("Failed to open template %s", input)
		}
		defer fp.Close()
		rd = fp
	}

	w := world.New(&world.Options{
		Logger:     log,
		Insecure:   insecure,
		LeftDelim:  leftDelim,
		RightDelim: rightDelim,
	})
	if vaultPrefix != "" {
		w.Vault().Prefix = vaultPrefix
		w.Azure().Prefix = azurePrefix
	}
	if vaultMapping != "" {
		vaultMap, err := loadKeyMapping(vaultMapping)
		if err != nil {
			log.WithError(err).Fatalf("Failed to load vault mapping file")
		}
		azureMap, err := loadKeyMapping(azureMapping)
		if err != nil {
			log.WithError(err).Fatalf("Failed to load azure mapping file")
		}
		w.Vault().KeyMapping = vaultMap
		w.Azure().KeyMapping = azureMap
	}
	wd, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("Failed to determine current working directory.")
	}
	d, err := world.LoadData(data, wd)
	if err != nil {
		log.WithError(err).Fatalf("Failed to load data")
	}
	w.Data = d

	output := bytes.Buffer{}
	if err := w.Render(&output, rd); err != nil {
		log.WithError(err).Fatal("Failed to render")
	}
	if outputFile == "" {
		io.Copy(os.Stdout, &output)
	} else {
		fp, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.WithError(err).Fatal("Failed to open output file")
		}
		defer fp.Close()
		if _, err := io.Copy(fp, &output); err != nil {
			log.WithError(err).Fatal("Failed to write to output file")
		}
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
