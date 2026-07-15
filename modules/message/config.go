package message

type Config struct {
	MaxLength    int
	DefaultLimit int
}

func DefaultConfig() Config {
	return Config{MaxLength: 4000, DefaultLimit: 50}
}
