package uploader

type Config struct {
	Address    string
	Host       string
	Port       int
	Login      string
	Password   string
	APIKey     string
	Path       string
	Extensions []string
}
