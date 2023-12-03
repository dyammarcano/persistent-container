//go:build generate && prod

package main

import (
	"github.com/caarlos0/log"
	"os"
	"os/exec"
)

//go:generate go run gen-prod.go

// to execute this file, run `go generate -prod ./...` from the root directory of this project

func main() {
	//change directory
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}

	// build vue project in production mode
	if err := exec.Command("npm", "run", "build").Run(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("project built successfully")
}
