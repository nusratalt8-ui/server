package dbutil

type Config struct {
	Path         string
	SchemaPath   string
	MaxOpenConns int
	BusyTimeout  int
}

func DefaultConfig() Config {
	return Config{
		Path:         "data/agentmanager.db",
		SchemaPath:   "data/schema.sql",
		MaxOpenConns: 1,
		BusyTimeout:  5000,
	}
}
