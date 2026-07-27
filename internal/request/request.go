package request

type Request interface {
	Send() error
}

type RequestMeta struct {
	Name     string `toml:"name"`
	Protocol string `toml:"protocol"`
}
