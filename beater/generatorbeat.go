package beater

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/urso/generatorbeat/config"
)

type Generatorbeat struct {
	wg   sync.WaitGroup
	done chan struct{}

	client publisher.Client
	worker []*worker
}

type worker struct {
	done <-chan struct{}
	name string
	gen  func() common.MapStr
}

type generatorFunc func() common.MapStr

type generatorFactory func(cfg *common.Config) ([]generatorFunc, error)

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

	// Load beater beatConfig
	cfg := config.Config{}
	err := b.RawConfig.Unpack(&cfg)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	for name, cfg := range cfg.Generatorbeat.Generators {
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
	bt.client = b.Publisher.Connect()

	return nil
}

func (bt *Generatorbeat) Run(b *beat.Beat) error {
	for _, w := range bt.worker {
		bt.wg.Add(1)
		go func(worker *worker) {
			defer bt.wg.Done()
			worker.run(bt.client)
		}(w)
	}

	bt.wg.Wait()
	return nil
}

func (bt *Generatorbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Generatorbeat) Stop() {
	bt.client.Close()
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
		Worker int `config:"worker" validate:"min=1"`
		Repeat int `config:"repeat" validate:"min=1"`
	}{
		Worker: 1,
		Repeat: 1,
	}
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
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
		return func() common.MapStr {
			line := genLine()
			off := offset
			offset += uint64(len(line))
			return common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"type":       "filebeat",
				"message":    line,
				"offset":     off,
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
