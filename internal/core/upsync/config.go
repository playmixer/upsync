package upsync

type Config struct {
	WorkerPoolCount int `env:"UPSYNC_WORKER_COUNT" envDefault:"1"`
}
