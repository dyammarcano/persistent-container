//go:build generate

package main

import (
	"github.com/caarlos0/log"
	"os"
	"os/exec"
)

//go:generate go run gen.go

func main() {
	//change directory
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}

	log.Info("build vue project")

	// build vue project
	if err := exec.Command("npm", "run", "build").Run(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("project built successfully")
}
