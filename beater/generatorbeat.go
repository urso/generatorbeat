package beater

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"

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
	gen  func() common.MapStr
}

type generatorFunc func() common.MapStr

type generatorFactory func(cfg config.WorkerConfig) ([]generatorFunc, error)

// Creates beater
func New() *Generatorbeat {
	return &Generatorbeat{
		done: make(chan struct{}),
	}
}

var generators = map[string]generatorFactory{
	"filebeat":   genFilebeat,
	"topbeat":    genTopbeat,
	"packetbeat": genPacketbeat,
}

/// *** Beater interface methods ***///

func (bt *Generatorbeat) Config(b *beat.Beat) error {

	cfg := &config.GeneratorbeatConfig{}
	err := cfgfile.Read(&cfg, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	for name, cfg := range cfg.Generators {
		factory, ok := generators[name]
		if !ok {
			return fmt.Errorf("Unknown generator: %v", name)
		}

		generators, err := factory(cfg)
		if err != nil {
			return err
		}

		for _, gen := range generators {
			bt.worker = append(bt.worker, &worker{
				done: bt.done,
				gen:  gen,
			})
		}
	}

	return nil
}

func (bt *Generatorbeat) Setup(b *beat.Beat) error {
	return nil
}

func (bt *Generatorbeat) Run(b *beat.Beat) error {
	for _, w := range bt.worker {
		bt.wg.Add(1)
		go func(worker *worker) {
			defer bt.wg.Done()
			worker.run(b.Events)
		}(w)
	}

	bt.wg.Wait()
	return nil
}

func (bt *Generatorbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Generatorbeat) Stop() {
	close(bt.done)
}

func (w *worker) run(client publisher.Client) {
	for w.running() {
		event := w.gen()
		client.PublishEvent(event)
	}
}

func (w *worker) running() bool {
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}

func genFilebeat(cfg config.WorkerConfig) ([]generatorFunc, error) {
	text := strings.Split(`Lorem ipsum dolor sit amet, consetetur sadipscing elitr,
sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat,
sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet
clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod
tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At
vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren,
no sea takimata sanctus est Lorem ipsum dolor sit amet.`, "\n")

	makeGen := func() generatorFunc {
		i := 0
		return func() common.MapStr {
			line := text[i]
			i++
			if i >= len(text) {
				i = 0
			}

			return common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"type":       "filebeat",
				"message":    line,
				"offset":     i,
			}
		}
	}

	n := 1
	if cfg.Worker > n {
		n = cfg.Worker
	}

	var generators []generatorFunc
	for i := 0; i < n; i++ {
		generators = append(generators, makeGen())
	}
	return generators, nil
}

func genTopbeat(cfg config.WorkerConfig) ([]generatorFunc, error) {
	return nil, errors.New("topbeat mode not yet implemented")
}

func genPacketbeat(cfg config.WorkerConfig) ([]generatorFunc, error) {
	return nil, errors.New("packetbeat mode not yet implemented")
}
