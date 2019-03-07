package main

import (
	"os"

	"github.com/elastic/beats/libbeat/cmd"
	"github.com/elastic/beats/libbeat/cmd/instance"

	"github.com/urso/generatorbeat/beater"
)

const name = "generatorbeat"

func main() {
	rootCmd := cmd.GenRootCmdWithSettings(beater.New, instance.Settings{
		Name: "generatorbeat",
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
