package main

import (
	"os"

	"github.com/elastic/beats/libbeat/cmd"

	"github.com/urso/generatorbeat/beater"
)

const name = "generatorbeat"

func main() {
	if err := cmd.GenRootCmd(name, "", beater.New).Execute(); err != nil {
		os.Exit(1)
	}
}
