package beater

import (
	"bytes"
	"compress/bzip2"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/urso/generatorbeat/config"
)

type Generatorbeat struct {
	wg   sync.WaitGroup
	done chan struct{}

	worker []*worker
}

type worker struct {
	done <-chan struct{}
	name string
	gen  generatorFunc
}

type generatorFunc func() beat.Event

type generatorFactory func(cfg *common.Config) ([]generatorFunc, error)

var (
	max = flag.Int("max", -1, "Quit after pushing 'max' events")
)

var generators = map[string]generatorFactory{
	"filebeat":   genFilebeat,
	"topbeat":    genTopbeat,
	"packetbeat": genPacketbeat,
}

// Creates beater
func New(b *beat.Beat, rawConfig *common.Config) (beat.Beater, error) {
	cfg := config.DefaultConfig
	err := rawConfig.Unpack(&cfg)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	done := make(chan struct{})

	workers := []*worker{}
	for name, cfg := range cfg.Generators {
		factory, ok := generators[name]
		if !ok {
			return nil, fmt.Errorf("Unknown generator: %v", name)
		}

		generators, err := factory(cfg)
		if err != nil {
			return nil, err
		}

		for _, gen := range generators {
			workers = append(workers, &worker{
				done: done,
				gen:  gen,
			})
		}
	}

	return &Generatorbeat{
		done:   done,
		worker: workers,
	}, nil
}

func (bt *Generatorbeat) Run(b *beat.Beat) error {
	for _, w := range bt.worker {
		bt.wg.Add(1)

		client, err := b.Publisher.Connect()
		if err != nil {
			return err
		}

		go func(worker *worker) {
			defer bt.wg.Done()
			worker.run(client)
		}(w)
	}

	bt.wg.Wait()
	return nil
}

func (bt *Generatorbeat) Stop() {
	close(bt.done)
}

func (w *worker) run(client beat.Client) {
	go func() {
		<-w.done
		client.Close()
	}()

	if *max > 0 {
		m := *max
		for w.running() && m >= 0 {
			event := w.gen()
			client.Publish(event)
			m--
		}
	} else {
		for w.running() {
			event := w.gen()
			client.Publish(event)
		}
	}

}

func (w *worker) Stop() {

}

func (w *worker) running() bool {
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}

func genFilebeat(cfg *common.Config) ([]generatorFunc, error) {
	text := strings.Split(`Lorem ipsum dolor sit amet, consetetur sadipscing elitr,
sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat,
sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet
clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod
tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At
vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren,
no sea takimata sanctus est Lorem ipsum dolor sit amet.`, "\n")

	config := struct {
		Worker     int    `config:"worker" validate:"min=1"`
		Repeat     int    `config:"repeat" validate:"min=1"`
		SampleFile string `config:"sample_file"`
	}{
		Worker: 1,
		Repeat: 1,
	}
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	if config.SampleFile != "" {
		logp.Info("Read sample file: %v", config.SampleFile)

		if filepath.Ext(config.SampleFile) == ".bz2" {
			f, err := os.Open(config.SampleFile)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			content, err := ioutil.ReadAll(bzip2.NewReader(f))
			if err != nil {
				return nil, err
			}
			text = strings.Split(string(content), "\n")
		} else {
			content, err := ioutil.ReadFile(config.SampleFile)
			if err != nil {
				return nil, err
			}
			text = strings.Split(string(content), "\n")
		}
	}

	makeGenLine := func() func() string {
		i := 0

		nextLine := func() string {
			line := text[i]
			i++
			if i >= len(text) {
				i = 0
			}
			return line
		}

		return func() string {
			if config.Repeat == 1 {
				return nextLine()
			}

			buf := bytes.NewBuffer(nil)
			for j := 0; j < config.Repeat; j++ {
				buf.WriteString(nextLine())
			}
			return buf.String()
		}
	}

	makeGen := func() generatorFunc {
		genLine := makeGenLine()
		var offset uint64
		return func() beat.Event {
			line := genLine()
			off := offset
			offset += uint64(len(line))
			return beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":    "filebeat",
					"message": line,
					"offset":  off,
				},
			}
		}
	}

	var generators []generatorFunc
	for i := 0; i < config.Worker; i++ {
		generators = append(generators, makeGen())
	}
	return generators, nil
}

func genTopbeat(cfg *common.Config) ([]generatorFunc, error) {
	return nil, errors.New("topbeat mode not yet implemented")
}

func genPacketbeat(cfg *common.Config) ([]generatorFunc, error) {
	return nil, errors.New("packetbeat mode not yet implemented")
}
