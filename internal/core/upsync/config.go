package upsync

type Config struct {
	WorkerPoolCount int `env:"UPSYNC_WOKRER_COUNT" envDefault:"1"`
}
