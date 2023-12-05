//go:build generate && prod

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/caarlos0/log"
	"os"
	"strconv"
	"strings"
)

//go:generate go run gen.go

// Generate a master key
func generateKeys(size int, setZero bool) (string, error) {
	if setZero {
		masterKey := strings.Builder{}
		for i := 0; i < size; i++ {
			masterKey.WriteString("00")
		}
		return masterKey.String(), nil
	}

	masterKey := make([]byte, size)
	if _, err := rand.Read(masterKey); err != nil {
		return "", err
	}
	return hex.EncodeToString(masterKey), nil
}

func getSpaces(value, totalKeys int) string {
	strNumKeys := strconv.Itoa(totalKeys)
	if len(strNumKeys) == 4 {
		strNum := strconv.Itoa(value)
		if len(strNum) == 1 {
			return "    "
		}

		if len(strNum) == 2 {
			return "   "
		}

		if len(strNum) == 3 {
			return "  "
		}

		return " "
	}

	if len(strNumKeys) == 3 {
		strNum := strconv.Itoa(value)
		if len(strNum) == 1 {
			return "   "
		}

		if len(strNum) == 2 {
			return "  "
		}

		return " "
	}

	return " "
}

func main() {
	fileName := "generated_file.go"
	keysLength, keysNumber := 44, 512

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`// Package crypto this package was created by with 'go generate ./...' command

package cryptography

const (
	keysLenght = %d
)

var stringKeys = map[int]string{
`, keysLength))

	for i := 0; i < keysNumber; i++ {
		data, err := generateKeys(keysLength, false)
		if err != nil {
			panic(err)
		}
		builder.WriteString(fmt.Sprintf(`	%d:%s"%s",
`, i, getSpaces(i, keysNumber), data))
	}

	builder.WriteString(`}
`)

	//change directory
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}

	outputFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	if _, err = outputFile.WriteString(builder.String()); err != nil {
		panic(err)
	}

	log.Infof("generated key file: %s", fileName)
}
