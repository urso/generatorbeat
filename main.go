package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/urso/generatorbeat/beater"
)

func main() {
	err := beat.Run("generatorbeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
