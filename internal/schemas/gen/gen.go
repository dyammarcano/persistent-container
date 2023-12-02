//go:build generate && windows

package main

import (
	"github.com/caarlos0/log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:generate go run gen.go

func main() {
	log.Info("generating code...")

	if _, err := exec.LookPath("flatc"); err != nil {
		log.Errorf("flatc not found: %v", err)
	}

	if err := os.Chdir(".."); err != nil {
		log.Errorf("error changing directory: %v", err)
	}

	source, err := filepath.Abs(".")
	if err != nil {
		log.Errorf("error getting absolute path: %v", err)
	}

	var filePaths []string

	files, err := os.ReadDir(source)
	if err != nil {
		log.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".fbs" {
			filePaths = append(filePaths, filepath.Join(source, file.Name()))
		}
	}

	output, err := filepath.Abs("../")
	if err != nil {
		log.Errorf("error getting absolute path: %v", err)
	}

	for _, file := range filePaths {
		if err = processFile(file, output); err != nil {
			log.Errorf("error generating code: %v", err)
		}
	}
}

func processFile(source, output string) error {
	log.Infof("processing file: %s", source)
	return exec.Command("flatc", "--go", "--gen-object-api", "-o", output, source).Run()
}
