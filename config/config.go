// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Generatorbeat GeneratorbeatConfig
}

type GeneratorbeatConfig struct {
	Generators map[string]WorkerConfig
}

type WorkerConfig struct {
	Worker int
}
