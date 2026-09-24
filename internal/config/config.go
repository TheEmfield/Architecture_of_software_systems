package config

type Config struct {
	Logger    Logger    `yaml:"logger"`
	Simulator Simulator `yaml:"simulator"`
}

type Logger struct {
	Level  string `yaml:"level"  env:"LOG_LEVEL"  env-default:"info"`
	Format string `yaml:"format" env:"LOG_FORMAT" env-default:"text"`
}

type Simulator struct {
	NumDevices        int     `yaml:"num_devices"`
	NumSources        int     `yaml:"num_sources"`
	BufferCapacity    int     `yaml:"buffer_capacity"`
	MinInterval       float64 `yaml:"min_interval"`
	MaxInterval       float64 `yaml:"max_interval"`
	Lambda            float64 `yaml:"lambda"`
	MaxSimulationTime float64 `yaml:"max_simulation_time"`
}
