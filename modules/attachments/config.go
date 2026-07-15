package attachments

type Config struct {
	StoragePath string
	MaxSize     int64
	MasterKey   string
}

func DefaultConfig() Config {
	return Config{
		StoragePath: "data/attachments",
		MaxSize:     100 * 1024 * 1024,
	}
}
