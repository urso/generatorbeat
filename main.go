package main

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/urso/generatorbeat/beater"
)

func main() {
	beat.Run("generatorbeat", "", beater.New())
}
