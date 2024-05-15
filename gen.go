//go:build generate

package main

//go:generate go run gen.go

import (
	"github.com/dyammarcano/version"
)

func main() {
	ver, err := version.NewVersion()
	if err != nil {
		panic(err)
	}

	if err = ver.Generate(); err != nil {
		panic(err)
	}
}
