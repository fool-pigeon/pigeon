package request

type Response interface{}
type Request interface {
	Send() (Response, error)
}

type RequestMeta struct {
	Name     string `toml:"name"`
	Protocol string `toml:"protocol"`
}
